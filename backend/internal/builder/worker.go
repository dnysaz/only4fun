package builder

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"only4.fun/backend/internal/database"
	"only4.fun/backend/internal/dns"
)

// Worker memproses antrian deployment satu per satu (bisa beberapa goroutine).
type Worker struct {
	builder     *Builder
	queue       chan Job
	projects    *database.ProjectStore
	deployments *database.DeploymentStore
	dns         *dns.Client
}

func NewWorker(b *Builder, projects *database.ProjectStore, deployments *database.DeploymentStore, dnsClient *dns.Client) *Worker {
	return &Worker{
		builder:     b,
		queue:       make(chan Job, 64),
		projects:    projects,
		deployments: deployments,
		dns:         dnsClient,
	}
}

// Start menjalankan sejumlah goroutine pemroses.
func (w *Worker) Start(concurrency int) {
	if concurrency < 1 {
		concurrency = 1
	}
	for i := 0; i < concurrency; i++ {
		go w.loop()
	}
	log.Printf("build worker started (mode=%s, concurrency=%d)", w.builder.Mode(), concurrency)
}

// Enqueue memasukkan job ke antrian. Mengembalikan error bila antrian penuh.
func (w *Worker) Enqueue(job Job) error {
	select {
	case w.queue <- job:
		return nil
	default:
		return fmt.Errorf("the build queue is full, please try again later")
	}
}

func (w *Worker) loop() {
	for job := range w.queue {
		w.process(job)
	}
}

func (w *Worker) process(job Job) {
	var deployment *database.Deployment
	defer func() {
		if r := recover(); r != nil {
			log.Printf("worker panic on project %s: %v", job.ProjectID, r)
			_ = w.projects.UpdateStatus(job.ProjectID, "failed")
			// Tanpa ini record deployment tertinggal di status "building" dan
			// terminal build di dashboard terus menunggu selamanya.
			if deployment != nil {
				_ = w.deployments.UpdateStatus(deployment.ID, "failed", fmt.Sprintf("worker panic: %v", r))
			}
		}
	}()

	var err error
	deployment, err = w.deployments.Create(job.ProjectID, "building", "")
	if err != nil {
		log.Printf("failed to create the deployment record for project %s: %v", job.ProjectID, err)
	}
	if err := w.projects.UpdateStatus(job.ProjectID, "building"); err != nil {
		log.Printf("failed to update project status %s: %v", job.ProjectID, err)
	}

	var mu sync.Mutex
	var lines []string
	var lastFlush time.Time

	// persist menulis progres ke database supaya halaman project bisa
	// menampilkan log secara bertahap selagi build masih berjalan.
	persist := func() {
		mu.Lock()
		defer mu.Unlock()
		if deployment == nil {
			return
		}
		if err := w.deployments.UpdateStatus(deployment.ID, "building", strings.Join(lines, "\n")); err != nil {
			log.Printf("failed to save deployment progress %s: %v", deployment.ID, err)
		}
	}

	logf := func(line string) {
		mu.Lock()
		if line == "" || len(lines) >= 500 {
			mu.Unlock()
			return
		}
		lines = append(lines, fmt.Sprintf("%s %s", time.Now().Format("15:04:05"), line))
		// Batasi write ke database maksimal sekali per detik.
		due := time.Since(lastFlush) >= time.Second
		if due {
			lastFlush = time.Now()
		}
		mu.Unlock()

		if due {
			persist()
		}
	}

	ctx := context.Background()
	outDir, buildErr := w.builder.Build(ctx, job, logf)

	finalStatus := "success"
	if buildErr != nil {
		finalStatus = "failed"
		logf("ERROR: " + buildErr.Error())
	} else {
		logf("ok: " + outDir)
		// Simpan nama folder rilis agar deployment ini bisa di-rollback nanti.
		if deployment != nil {
			if release, releaseErr := CurrentReleaseName(outDir); releaseErr == nil {
				if err := w.deployments.SetRelease(deployment.ID, release); err != nil {
					log.Printf("failed to save the release for deployment %s: %v", deployment.ID, err)
				}
				deployment.Release = release
			}
		}

		// Daftarkan subdomain di Cloudflare agar situs bisa diakses publik.
		if w.dns != nil && w.dns.Enabled() {
			logf("dns: registering " + job.Subdomain)
			if dnsErr := w.dns.EnsureRecord(ctx, job.Subdomain); dnsErr != nil {
				finalStatus = "failed"
				logf("dns error: " + dnsErr.Error())
			} else {
				logf("dns: ok")
			}
		} else {
			logf("dns: Cloudflare integration is disabled, skipping the subdomain record")
		}
	}

	logText := strings.Join(lines, "\n")
	if deployment != nil {
		if err := w.deployments.UpdateStatus(deployment.ID, finalStatus, logText); err != nil {
			log.Printf("failed to save the deployment log %s: %v", deployment.ID, err)
		}
	}

	projectStatus := "deployed"
	if finalStatus == "failed" {
		projectStatus = "failed"
	}
	if err := w.projects.UpdateStatus(job.ProjectID, projectStatus); err != nil {
		log.Printf("failed to update project status %s: %v", job.ProjectID, err)
	}
	log.Printf("deployment for project %s finished: %s", job.ProjectID, projectStatus)
}
