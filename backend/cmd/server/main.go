package main

import (
	"flag"
	"log"

	"only4.fun/backend/internal/builder"
	"only4.fun/backend/internal/config"
	"only4.fun/backend/internal/database"
	"only4.fun/backend/internal/dns"
	"only4.fun/backend/internal/secure"
	"only4.fun/backend/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Kunci enkripsi untuk token GitHub dan hashing session.
	if err := secure.LoadFromEnv(); err != nil {
		log.Fatalf("Failed to prepare the encryption key: %v", err)
	}

	// Inisialisasi database (SQLite)
	db, err := database.Init(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize the database: %v", err)
	}
	defer db.Close()

	// Jalankan migrasi schema
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to migrate the database: %v", err)
	}

	// Siapkan pipeline build (clone repo, build, publish).
	workers := flag.Int("workers", 1, "jumlah build yang dijalankan paralel")
	flag.Parse()

	stores := server.NewStores(db)
	buildEngine := builder.New(cfg.Build)
	if buildEngine.Mode() == "local" && !cfg.Build.AllowLocal {
		log.Printf("WARNING: Docker is unavailable and local builds are not allowed; builds will fail. Set BUILDER_ALLOW_LOCAL=true only if you trust every repository that can be deployed.")
	}
	dnsClient := dns.New(cfg.Cloudflare)
	if dnsClient.Enabled() {
		log.Printf("integrasi Cloudflare aktif untuk zona %s (target %s, proxy=%v)", cfg.Cloudflare.Domain, cfg.Cloudflare.Target, cfg.Cloudflare.Proxy)
	} else {
		log.Printf("integrasi Cloudflare tidak aktif: set CLOUDFLARE_API_KEY, CLOUDFLARE_ZONE_ID, dan CLOUDFLARE_TARGET untuk DNS otomatis")
	}
	worker := builder.NewWorker(buildEngine, stores.Project, stores.Deployment, dnsClient)
	worker.Start(*workers)

	// Buat dan jalankan HTTP server
	srv := server.New(cfg, db, worker)
	log.Printf("only4.fun backend listening on %s", cfg.Server.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
