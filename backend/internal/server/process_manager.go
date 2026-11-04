package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// ProcessManager mengelola proses Node.js untuk SSR projects.
type ProcessManager struct {
	mu        sync.Mutex
	processes map[string]*Process
	nextPort  int
	maxProcs  int
}

// Process merepresentasikan satu proses Node.js untuk satu project.
type Process struct {
	Subdomain  string
	Port       int
	Cmd        *exec.Cmd
	StartedAt  time.Time
	LastAccess time.Time
}

// NewProcessManager membuat process manager baru.
func NewProcessManager() *ProcessManager {
	pm := &ProcessManager{
		processes: make(map[string]*Process),
		nextPort:  4000,
		maxProcs:  50,
	}
	// Jalankan idle cleanup setiap 5 menit
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			pm.StopIdle(10 * time.Minute)
		}
	}()
	return pm
}

// Start memulai proses Node.js untuk subdomain tertentu.
// serverEntry adalah path ke file entry point (misal: server.js).
//
// Bila proses lama sudah usang (build baru menulis ulang entry point), proses
// baru dinyalakan di port lain dan baru setelah siap port aktif ditukar, lalu
// proses lama dimatikan. Dengan begitu tidak ada jeda 502 selama redeploy.
func (pm *ProcessManager) Start(subdomain, serverEntry string) (int, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	old := pm.processes[subdomain]
	if old != nil && !isProcessAlive(old) {
		delete(pm.processes, subdomain)
		old = nil
	}

	if old != nil {
		// Build baru menulis ulang file entry point. Jika file lebih baru
		// dari proses yang berjalan, restart agar versi terbaru dipakai
		// (tanpa ini redeploy SSR terus menyajikan kode lama).
		stale := false
		if info, err := os.Stat(serverEntry); err == nil && info.ModTime().After(old.StartedAt) {
			stale = true
		}
		if !stale {
			old.LastAccess = time.Now()
			return old.Port, nil
		}
	}

	// Proses lama untuk subdomain yang sama akan digantikan, jadi tidak dihitung.
	if old == nil && len(pm.processes) >= pm.maxProcs {
		return 0, fmt.Errorf("too many running servers, please try again later")
	}

	p, err := pm.startProcess(subdomain, serverEntry)
	if err != nil {
		// Jangan matikan versi lama bila versi baru gagal siap; lebih baik
		// tetap menyajikan kode lama daripada 503.
		if old != nil {
			log.Printf("SSR restart for %s failed (%v), keeping port %d", subdomain, err, old.Port)
			old.LastAccess = time.Now()
			return old.Port, nil
		}
		delete(pm.processes, subdomain)
		return 0, err
	}

	pm.processes[subdomain] = p
	if old != nil {
		if old.Cmd.Process != nil {
			_ = old.Cmd.Process.Kill()
		}
		log.Printf("SSR server swapped for %s: port %d -> %d", subdomain, old.Port, p.Port)
	}
	log.Printf("SSR server started for %s on port %d", subdomain, p.Port)
	return p.Port, nil
}

// startProcess menyalakan satu proses Node.js dan menunggu port-nya siap.
// Pemanggil harus memegang pm.mu.
func (pm *ProcessManager) startProcess(subdomain, serverEntry string) (*Process, error) {
	// Alokasi port
	port := pm.nextPort
	pm.nextPort++
	if pm.nextPort > 5000 {
		pm.nextPort = 4000
	}

	// Buat command
	cmd := exec.Command("node", serverEntry)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", port))
	cmd.Dir = filepath.Dir(serverEntry)

	// Redirect output ke log
	var logFile *os.File
	if lf, err := os.OpenFile(filepath.Join(os.TempDir(), fmt.Sprintf("ssr-%s.log", subdomain)), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		logFile = lf
		cmd.Stdout = lf
		cmd.Stderr = lf
	}

	if err := cmd.Start(); err != nil {
		if logFile != nil {
			_ = logFile.Close()
		}
		return nil, fmt.Errorf("failed to start the server: %w", err)
	}

	// Reap proses saat keluar (termasuk setelah Kill) dan tutup file log.
	// Tanpa Wait, proses yang dimatikan tetap menjadi zombie dan file log
	// tetap terbuka sampai server utama berhenti.
	go func() {
		_ = cmd.Wait()
		if logFile != nil {
			_ = logFile.Close()
		}
	}()

	// Tunggu server siap (max 10 detik)
	ready := false
	for i := 0; i < 50; i++ {
		time.Sleep(200 * time.Millisecond)
		if isPortOpen(port) {
			ready = true
			break
		}
	}
	if !ready {
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("the server did not become ready within 10 seconds")
	}

	return &Process{
		Subdomain:  subdomain,
		Port:       port,
		Cmd:        cmd,
		StartedAt:  time.Now(),
		LastAccess: time.Now(),
	}, nil
}

// Stop menghentikan proses Node.js untuk subdomain tertentu.
func (pm *ProcessManager) Stop(subdomain string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if p, ok := pm.processes[subdomain]; ok {
		if p.Cmd.Process != nil {
			p.Cmd.Process.Kill()
		}
		delete(pm.processes, subdomain)
		log.Printf("SSR server stopped for %s", subdomain)
	}
}

// GetPort mengembalikan port server untuk subdomain, atau 0 jika tidak berjalan.
func (pm *ProcessManager) GetPort(subdomain string) int {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if p, ok := pm.processes[subdomain]; ok {
		if isProcessAlive(p) {
			p.LastAccess = time.Now()
			return p.Port
		}
		delete(pm.processes, subdomain)
	}
	return 0
}

// StopIdle menghentikan proses yang tidak diakses selama lebih dari idleTimeout.
func (pm *ProcessManager) StopIdle(idleTimeout time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	now := time.Now()
	for subdomain, p := range pm.processes {
		if now.Sub(p.LastAccess) > idleTimeout {
			if p.Cmd.Process != nil {
				p.Cmd.Process.Kill()
			}
			delete(pm.processes, subdomain)
			log.Printf("SSR server stopped (idle) for %s", subdomain)
		}
	}
}

// StopAll menghentikan semua proses.
func (pm *ProcessManager) StopAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for subdomain, p := range pm.processes {
		if p.Cmd.Process != nil {
			p.Cmd.Process.Kill()
		}
		delete(pm.processes, subdomain)
	}
}

// StopAllGraceful menghentikan semua proses Node.js secara rapi (SIGTERM),
// lalu menunggu maksimal graceTimeout sebelum dipaksa matikan.
func (pm *ProcessManager) StopAllGraceful(graceTimeout time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for subdomain, p := range pm.processes {
		if p.Cmd.Process == nil {
			delete(pm.processes, subdomain)
			continue
		}
		// Kirim SIGTERM agar Node.js bisa menyelesaikan request yang sedang berjalan.
		_ = p.Cmd.Process.Signal(syscall.SIGTERM)
		log.Printf("SSR server graceful shutdown started for %s", subdomain)
	}

	if graceTimeout <= 0 {
		return
	}

	// Tunggu hingga semua proses keluar atau timeout.
	done := make(chan struct{})
	go func() {
		for _, p := range pm.processes {
			_ = p.Cmd.Wait()
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(graceTimeout):
		// Paksa matikan yang tersisa.
		for subdomain, p := range pm.processes {
			if p.Cmd.Process != nil {
				_ = p.Cmd.Process.Kill()
				_ = p.Cmd.Wait()
			}
			delete(pm.processes, subdomain)
		}
	}
}

// RegisterSignalHandler mendaftarkan handler SIGTERM/SIGINT agar proses
// Node.js (SSR) dihentikan secara rapi saat backend dihentikan (misal
// `systemctl restart`). Setelah selesai, os.Exit(0) dipanggil.
func (pm *ProcessManager) RegisterSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-c
		log.Println("received shutdown signal, stopping SSR servers gracefully...")
		pm.StopAllGraceful(10 * time.Second)
		os.Exit(0)
	}()
}

func isProcessAlive(p *Process) bool {
	if p.Cmd.Process == nil {
		return false
	}
	// Signal 0 checks if process exists without killing it.
	err := p.Cmd.Process.Signal(syscall.Signal(0))
	return err == nil
}

func isPortOpen(port int) bool {
	// Coba koneksi TCP ke port
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(port), 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
