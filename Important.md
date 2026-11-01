# Important — Production Notes

## Catatan untuk Produksi

- **Set `SESSION_SECRET` env var** (jangan pakai default key `only4.fun-dev-insecure-default-key`). Tanpa ini, enkripsi token GitHub dan env vars tidak aman.
- **Proteksi file `data/only4.fun.db`** — set file permission (chmod 600), backup rutin, jangan expose ke publik.
- **Batasi jumlah proses SSR** — `maxProcs = 50` di `ProcessManager`. Monitor dan sesuaikan jika perlu.
- **Monitor resource usage (memory, CPU)** untuk proses Node.js SSR yang berjalan. Idle cleanup otomatis setiap 5 menit (matikan proses idle > 10 menit).
- **Pertimbangkan rate limiting** untuk endpoint `/api/check-subdomain` (saat ini tidak ada rate limit khusus).

## Arsitektur

- **Backend**: Go (Gin) + SQLite (`data/only4.fun.db`)
- **Frontend**: Next.js (port 3000)
- **Build**: Clone repo → detect framework → install deps → build → publish ke `data/sites/<subdomain>/`
- **SSR**: Node.js process per project (port 4000-5000), reverse proxy via Go
- **Static**: Serve file dari `data/sites/<subdomain>/`

## Keamanan

- Env vars dienkripsi AES-256-GCM (key dari `SESSION_SECRET`)
- `EnvVars` field punya `json:"-"` → tidak pernah di-serialize ke response
- `GetEnv` endpoint hanya return keys, bukan values
- Session cookie httpOnly, session ID di-hash (SHA-256)
- Auth middleware pada semua `/api/*` routes
- Path traversal protection di SiteHandler
- SQL injection safe (parameterized queries via sqlx)
- XSS/CSRF/SQL injection protection di frontend (`lib/check-word.ts`)

## Environment Variables

| Var | Default | Keterangan |
|-----|---------|------------|
| `SESSION_SECRET` | (insecure default) | **WAJIB di-set di production** |
| `SITES_DIR` | `./data/sites` | Folder hasil build |
| `CLOUDFLARE_API_KEY` | (kosong) | Untuk DNS otomatis (opsional) |
| `CLOUDFLARE_ZONE_ID` | (kosong) | Untuk DNS otomatis (opsional) |
| `CLOUDFLARE_TARGET` | (kosong) | Untuk DNS otomatis (opsional) |

## Deployment

Lihat `deploy.md` untuk panduan lengkap deploy ke VPS + Cloudflare, termasuk:
- `deploy.sh` script
- Systemd services
- Nginx reverse proxy config
- Cloudflare DNS & SSL setup
- Wildcard SSL (otomatis, tidak perlu cert per subdomain)

- Frontend: `npm run dev -- --port 3000`
- Backend: `go run ./cmd/server` (port 8080)
- Health check: `http://localhost:8080/health`

## Build Commands

- Backend: `cd backend && go build ./...`
- Frontend: `cd frontend && npx tsc --noEmit`
- Jangan jalankan `next build` saat `next dev` aktif (bisa corrupt `.next`)
