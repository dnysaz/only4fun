package server

import (
	"context"
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

	"only4.fun/backend/internal/execenv"
)

const (
	minSSRPort = 4000
	maxSSRPort = 5000
)

// ProcessManager mengelola proses Node.js untuk SSR projects.
type ProcessManager struct {
	mu        sync.Mutex
	processes map[string]*Process
	usedPorts map[int]bool
	nextPort  int
	maxProcs  int
	// keyLocks serializes start/stop per subdomain. The global mu is only held
	// for short map/port updates, never during the 10s readiness wait.
	keyLocks sync.Map
}

func (pm *ProcessManager) keyLock(subdomain string) *sync.Mutex {
	value, _ := pm.keyLocks.LoadOrStore(subdomain, &sync.Mutex{})
	return value.(*sync.Mutex)
}

// Process merepresentasikan satu proses Node.js untuk satu project.
type Process struct {
	Subdomain  string
	Port       int
	Cmd        *exec.Cmd
	StartedAt  time.Time
	LastAccess time.Time
	// done ditutup oleh goroutine Wait setelah proses benar-benar keluar.
	done chan struct{}
}

// NewProcessManager membuat process manager baru.
func NewProcessManager() *ProcessManager {
	pm := &ProcessManager{
		processes: make(map[string]*Process),
		usedPorts: make(map[int]bool),
		nextPort:  minSSRPort,
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

// allocatePort mencari port SSR yang benar-benar bebas. Pemanggil harus
// memegang pm.mu. Selain memeriksa daftar internal, port dicoba di-bind agar
// tidak mengambil port yang dipakai proses lain di mesin yang sama.
func (pm *ProcessManager) allocatePort() (int, error) {
	for i := 0; i <= maxSSRPort-minSSRPort; i++ {
		port := pm.nextPort
		pm.nextPort++
		if pm.nextPort > maxSSRPort {
			pm.nextPort = minSSRPort
		}
		if pm.usedPorts[port] {
			continue
		}
		ln, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		if err != nil {
			continue
		}
		_ = ln.Close()
		pm.usedPorts[port] = true
		return port, nil
	}
	return 0, fmt.Errorf("no free SSR port available")
}

func (pm *ProcessManager) releasePort(port int) {
	delete(pm.usedPorts, port)
}

// forget menghapus entri proses dan membebaskan portnya. Pemanggil harus
// memegang pm.mu.
func (pm *ProcessManager) forget(subdomain string) {
	if p := pm.processes[subdomain]; p != nil {
		pm.releasePort(p.Port)
		delete(pm.processes, subdomain)
	}
}

func killProcess(p *Process) {
	if p != nil && p.Cmd != nil && p.Cmd.Process != nil {
		_ = p.Cmd.Process.Kill()
	}
}

// Start memulai proses Node.js untuk subdomain tertentu.
// serverEntry adalah path ke file entry point (misal: server.js).
//
// Bila proses lama sudah usang (build baru menulis ulang entry point), proses
// baru dinyalakan di port lain dan baru setelah siap port aktif ditukar, lalu
// proses lama dimatikan. Dengan begitu tidak ada jeda 502 selama redeploy.
func (pm *ProcessManager) Start(subdomain, serverEntry string) (int, error) {
	// Startup bisa memakan waktu 10 detik. Kunci global hanya dipakai untuk
	// membaca/menulis state bersama, bukan saat menunggu proses siap.
	lock := pm.keyLock(subdomain)
	lock.Lock()
	defer lock.Unlock()

	pm.mu.Lock()
	old := pm.processes[subdomain]
	if old != nil && !isProcessAlive(old) {
		pm.forget(subdomain)
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
			pm.mu.Unlock()
			return old.Port, nil
		}
	}

	// Proses lama untuk subdomain yang sama akan digantikan, jadi tidak dihitung.
	if old == nil && len(pm.processes) >= pm.maxProcs {
		pm.mu.Unlock()
		return 0, fmt.Errorf("too many running servers, please try again later")
	}
	pm.mu.Unlock()

	p, err := pm.startProcess(subdomain, serverEntry)
	if err != nil {
		// Jangan matikan versi lama bila versi baru gagal siap; lebih baik
		// tetap menyajikan kode lama daripada 503.
		if old != nil {
			pm.mu.Lock()
			old.LastAccess = time.Now()
			pm.mu.Unlock()
			log.Printf("SSR restart for %s failed (%v), keeping port %d", subdomain, err, old.Port)
			return old.Port, nil
		}
		return 0, err
	}

	pm.mu.Lock()
	if old != nil {
		// Port lama dibebaskan sebelum entri ditimpa agar tidak bocor.
		pm.forget(subdomain)
		killProcess(old)
		log.Printf("SSR server swapped for %s: port %d -> %d", subdomain, old.Port, p.Port)
	}
	pm.processes[subdomain] = p
	pm.mu.Unlock()
	log.Printf("SSR server started for %s on port %d", subdomain, p.Port)
	return p.Port, nil
}

// ssrEnv memberi proses Node.js hanya environment minimum. PORT dibedakan per
// proses, sedangkan alamat bind dipaksa ke loopback: framework user boleh saja
// mengabaikan PORT, tetapi tidak boleh membuka port ke jaringan publik.
func ssrEnv(port int) []string {
	env := execenv.Base("")
	return append(env,
		fmt.Sprintf("PORT=%d", port),
		"HOST=127.0.0.1",
		"BIND=127.0.0.1",
		"LISTEN_ADDRESS=127.0.0.1",
	)
}

// startProcess menyalakan satu proses Node.js dan menunggu port-nya siap.
// Kunci global hanya dipakai saat mencatat/membebaskan port, bukan saat
// menunggu readiness, supaya subdomain lain tetap bisa dilayani.
func (pm *ProcessManager) startProcess(subdomain, serverEntry string) (*Process, error) {
	// Alokasi port yang benar-benar bebas.
	pm.mu.Lock()
	port, err := pm.allocatePort()
	pm.mu.Unlock()
	if err != nil {
		return nil, err
	}

	// Buat command. SSR menjalankan source milik user, jadi jangan pernah
	// meneruskan environment server yang berisi secret.
	if err := execenv.EnsureHome(""); err != nil {
		return nil, fmt.Errorf("failed to prepare SSR home: %w", err)
	}
	cmd := exec.Command("node", serverEntry)
	cmd.Env = ssrEnv(port)
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
		pm.mu.Lock()
		pm.releasePort(port)
		pm.mu.Unlock()
		return nil, fmt.Errorf("failed to start the server: %w", err)
	}

	// Reap proses saat keluar (termasuk setelah Kill) dan tutup file log.
	// Satu Wait hanya boleh dipanggil sekali; pemanggil lain menunggu channel
	// done instead of memanggil Cmd.Wait() untuk kedua kalinya.
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		if logFile != nil {
			_ = logFile.Close()
		}
		close(done)
	}()

	// Tunggu server siap (max 10 detik). Port harus benar-benar dibuka oleh
	// proses ini, bukan proses lain yang kebetulan memakai nomor yang sama.
	proc := &Process{Cmd: cmd}
	ready := false
	for i := 0; i < 50; i++ {
		time.Sleep(200 * time.Millisecond)
		if !isProcessAlive(proc) {
			pm.mu.Lock()
			pm.releasePort(port)
			pm.mu.Unlock()
			return nil, fmt.Errorf("the server exited during startup; check /tmp/ssr-%s.log", subdomain)
		}
		if isPortOpen(port) {
			ready = true
			break
		}
	}
	if !ready {
		killProcess(proc)
		pm.mu.Lock()
		pm.releasePort(port)
		pm.mu.Unlock()
		return nil, fmt.Errorf("the server did not become ready within 10 seconds")
	}

	return &Process{
		Subdomain:  subdomain,
		Port:       port,
		Cmd:        cmd,
		StartedAt:  time.Now(),
		LastAccess: time.Now(),
		done:       done,
	}, nil
}

// Stop menghentikan proses Node.js untuk subdomain tertentu.
func (pm *ProcessManager) Stop(subdomain string) {
	// Serialisasi dengan Start untuk subdomain yang sama.
	lock := pm.keyLock(subdomain)
	lock.Lock()
	defer lock.Unlock()

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if p := pm.processes[subdomain]; p != nil {
		killProcess(p)
		pm.forget(subdomain)
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
		pm.forget(subdomain)
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
			killProcess(p)
			pm.forget(subdomain)
			log.Printf("SSR server stopped (idle) for %s", subdomain)
		}
	}
}

// StopAll menghentikan semua proses.
func (pm *ProcessManager) StopAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for subdomain, p := range pm.processes {
		killProcess(p)
		pm.forget(subdomain)
	}
}

// StopAllGraceful menghentikan semua proses Node.js secara rapi (SIGTERM),
// lalu menunggu maksimal graceTimeout sebelum dipaksa matikan.
func (pm *ProcessManager) StopAllGraceful(graceTimeout time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	type entry struct {
		subdomain string
		process   *Process
	}
	entries := make([]entry, 0, len(pm.processes))
	for subdomain, p := range pm.processes {
		if p.Cmd == nil || p.Cmd.Process == nil {
			pm.forget(subdomain)
			continue
		}
		// Kirim SIGTERM agar Node.js bisa menyelesaikan request yang sedang berjalan.
		_ = p.Cmd.Process.Signal(syscall.SIGTERM)
		log.Printf("SSR server graceful shutdown started for %s", subdomain)
		entries = append(entries, entry{subdomain: subdomain, process: p})
	}
	if len(entries) == 0 {
		return
	}
	if graceTimeout <= 0 {
		graceTimeout = 10 * time.Second
	}

	deadline := time.NewTimer(graceTimeout)
	defer deadline.Stop()

	for _, e := range entries {
		select {
		case <-e.process.done:
			pm.forget(e.subdomain)
		case <-deadline.C:
			// Paksa matikan yang tersisa. Wait sudah dijalankan oleh goroutine
			// di startProcess, jadi cukup tunggu channel done-nya.
			for _, left := range entries {
				killProcess(left.process)
				<-left.process.done
				pm.forget(left.subdomain)
			}
			return
		}
	}
}

// RegisterSignalHandler mendaftarkan handler SIGTERM/SIGINT agar HTTP server
// berhenti menerima request baru, request berjalan selesai, lalu proses Node.js
// (SSR) dihentikan secara rapi (misal `systemctl restart`). Setelah selesai,
// os.Exit(0) dipanggil.
func (pm *ProcessManager) RegisterSignalHandler(srv *Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-c
		log.Println("received shutdown signal, draining HTTP requests...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if srv != nil {
			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("HTTP graceful shutdown: %v", err)
			}
		}
		log.Println("stopping SSR servers gracefully...")
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
