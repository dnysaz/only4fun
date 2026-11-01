# Deployment Guide — VPS + Cloudflare

## Arsitektur

```
User Browser
     │
     ▼
Cloudflare (SSL termination + WAF + DDoS protection + proxy)
     │  HTTPS (edge cert: *.only4.fun)
     ▼
VPS (Go backend :8080 + Next.js :3000)
     │  HTTP (internal, tidak exposed ke publik)
     ▼
data/sites/<subdomain>/  (static files)
     │
     ▼ (jika SSR)
Node.js process :4000-5000 (per project)
```

## SSL / Subdomain

**Tidak perlu mengatur SSL per subdomain.** Cloudflare Universal SSL otomatis mencakup:
- `only4.fun`
- `*.only4.fun` (semua subdomain)

Cukup set DNS record wildcard di Cloudflare:
```
Type: A
Name: *
Content: <IP VPS>
Proxy: Proxied (orange cloud) ← WAJIB agar SSL aktif
TTL: Auto
```

Dan satu record untuk root:
```
Type: A
Name: @ (atau only4.fun)
Content: <IP VPS>
Proxy: Proxied
```

### Mode SSL Cloudflare
Gunakan **Full (strict)**:
1. Cloudflare → SSL/TLS → Overview → Full (strict)
2. Generate Origin Certificate: SSL/TLS → Origin Server → Create Certificate
3. Install cert di VPS (nginx/caddy) atau Go backend

Atau gunakan **Flexible** (lebih mudah, tapi kurang aman):
- Cloudflare handle SSL ke user, VPS serve HTTP
- Tidak perlu cert di VPS

## deploy.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

# ─── Konfigurasi ──────────────────────────────────────────────────────────────
APP_DIR="/opt/only4.fun"
BACKEND_DIR="$APP_DIR/backend"
FRONTEND_DIR="$APP_DIR/frontend"
LOG_DIR="/var/log/only4.fun"
SERVICE_BACKEND="only4fun-backend"
SERVICE_FRONTEND="only4fun-frontend"

# ─── Warna output ─────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[deploy]${NC} $1"; }
warn() { echo -e "${YELLOW}[warn]${NC} $1"; }
err() { echo -e "${RED}[error]${NC} $1"; exit 1; }

# ─── Pre-flight checks ───────────────────────────────────────────────────────
log "Checking prerequisites..."
command -v go >/dev/null 2>&1 || err "Go not installed"
command -v node >/dev/null 2>&1 || err "Node.js not installed"
command -v npm >/dev/null 2>&1 || err "npm not installed"
command -v git >/dev/null 2>&1 || err "git not installed"

# ─── Pull latest code ────────────────────────────────────────────────────────
log "Pulling latest code..."
cd "$APP_DIR"
git pull origin main || err "git pull failed"

# ─── Backup database ─────────────────────────────────────────────────────────
log "Backing up database..."
DB_PATH="$BACKEND_DIR/data/only4.fun.db"
if [ -f "$DB_PATH" ]; then
    BACKUP_DIR="$BACKEND_DIR/data/backups"
    mkdir -p "$BACKUP_DIR"
    cp "$DB_PATH" "$BACKUP_DIR/only4.fun.db.$(date +%Y%m%d_%H%M%S)"
    # Keep only last 7 backups
    ls -t "$BACKUP_DIR"/only4.fun.db.* 2>/dev/null | tail -n +8 | xargs -r rm
    log "Database backed up"
fi

# ─── Build backend ───────────────────────────────────────────────────────────
log "Building backend..."
cd "$BACKEND_DIR"
go mod download
go build -o "$BACKEND_DIR/only4fun-server" ./cmd/server || err "backend build failed"
log "Backend built successfully"

# ─── Build frontend ──────────────────────────────────────────────────────────
log "Building frontend..."
cd "$FRONTEND_DIR"
npm ci --silent
npm run build || err "frontend build failed"
log "Frontend built successfully"

# ─── Run database migrations ─────────────────────────────────────────────────
log "Running migrations..."
# Migrations run automatically on server start (database.Migrate)
log "Migrations will run on server start"

# ─── Restart services ────────────────────────────────────────────────────────
log "Restarting services..."

# Backend
if systemctl is-active --quiet "$SERVICE_BACKEND"; then
    sudo systemctl restart "$SERVICE_BACKEND"
    log "Backend restarted"
else
    sudo systemctl start "$SERVICE_BACKEND"
    log "Backend started"
fi

# Frontend (Next.js production)
if systemctl is-active --quiet "$SERVICE_FRONTEND"; then
    sudo systemctl restart "$SERVICE_FRONTEND"
    log "Frontend restarted"
else
    sudo systemctl start "$SERVICE_FRONTEND"
    log "Frontend started"
fi

# ─── Health check ────────────────────────────────────────────────────────────
log "Waiting for backend health check..."
for i in {1..10}; do
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health | grep -q "200"; then
        log "Backend is healthy ✅"
        break
    fi
    if [ $i -eq 10 ]; then
        err "Backend health check failed after 10 attempts"
    fi
    sleep 2
done

log "Deployment complete! 🚀"
```

## Systemd Services

### Backend (`/etc/systemd/system/only4fun-backend.service`)
```ini
[Unit]
Description=only4.fun Backend
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/only4.fun/backend
Environment="SESSION_SECRET=your-secret-here"
Environment="SITES_DIR=/opt/only4.fun/backend/data/sites"
Environment="GIN_MODE=release"
ExecStart=/opt/only4.fun/backend/only4fun-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Frontend (`/etc/systemd/system/only4fun-frontend.service`)
```ini
[Unit]
Description=only4.fun Frontend
After=network.target only4fun-backend.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/only4.fun/frontend
Environment="NODE_ENV=production"
Environment="NEXT_PUBLIC_API_URL=https://api.only4.fun"
ExecStart=/usr/bin/npm start
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## Reverse Proxy (Nginx)

```nginx
# /etc/nginx/sites-available/only4.fun
server {
    listen 80;
    server_name only4.fun *.only4.fun;

    # Redirect HTTP to HTTPS (Cloudflare handles SSL, but just in case)
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name only4.fun *.only4.fun;

    # Cloudflare Origin Certificate
    ssl_certificate /etc/ssl/certs/cloudflare-origin.pem;
    ssl_certificate_key /etc/ssl/private/cloudflare-origin.key;

    # Frontend (Next.js)
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # Backend API
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Auth routes
    location /auth/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Health check
    location /health {
        proxy_pass http://127.0.0.1:8080;
    }
}
```

## Cloudflare DNS Setup

```
Type: A     Name: @     Content: <VPS_IP>     Proxy: Proxied
Type: A     Name: *     Content: <VPS_IP>     Proxy: Proxied
```

## Cloudflare SSL Settings

1. SSL/TLS → Overview → **Full (strict)**
2. SSL/TLS → Origin Server → Create Certificate
   - Hostnames: `*.only4.fun`, `only4.fun`
   - Validity: 15 years
   - Save cert and key to VPS
3. SSL/TLS → Edge Certificates → Always Use HTTPS: **On**
4. SSL/TLS → Edge Certificates → Minimum TLS Version: **1.2**

## Keamanan & Operasional

### Isolasi build

Build default dijalankan di container (`BUILDER_MODE=auto` → Docker bila tersedia). Bila Docker tidak
ada, build **gagal** dengan pesan jelas, karena menjalankan `npm install`/`build` repo pengguna langsung
di host berarti script repo berjalan dengan izin server (RCE).

Untuk pengembangan lokal di mesin yang dipercaya, izinkan secara eksplisit:

```bash
BUILDER_ALLOW_LOCAL=true
```

Jangan pernah menyalakan opsi ini di instance multi-user.

### HTTPS & cookie

Karena TLS berhenti di Cloudflare/nginx, backend mendeteksi HTTPS dari header proxy
(`X-Forwarded-Proto: https` atau `CF-Visitor: {"scheme":"https"}`) dan baru kemudian mengirim cookie
`Secure`. Pastikan Nginx meneruskan header tersebut (konfigurasi di atas sudah melakukannya):

```nginx
proxy_set_header X-Forwarded-Proto $scheme;
```

Gunakan mode SSL Cloudflare **Full (strict)** dengan Origin Certificate. Mode **Flexible** membuat hop
Cloudflare → VPS tetap HTTP polos sehingga cookie session dan token GitHub bisa disadap di jaringan.

### Struktur rilis & deploy atomik

```
data/sites/
├── <subdomain> -> .releases/<subdomain>/release-xxxx   (symlink)
└── .releases/<subdomain>/
    ├── release-xxxx   (rilis aktif)
    └── release-yyyy   (2 rilis cadangan terbaru)
```

Setiap deploy menyalin build ke folder rilis baru lalu mengganti symlink secara atomik, sehingga tidak
ada momen subdomain kehilangan folder aktif. Tiga rilis terbaru disimpan; sisanya dipangkas otomatis.
Saat project dihapus, symlink dan folder `.releases/<subdomain>` ikut dibersihkan.

### Custom domain & SSL

Domain kustom diatur per project dari halaman project (kolom `custom_domain`). Setelah disimpan:

1. Buat **CNAME** `domain-anda.com` → `<subdomain>.only4.fun` (atau A ke IP VPS).
2. Backend mendaftarkan hostname itu lewat API Cloudflare (`/zones/<zone>/custom_hostnames`) agar edge
   certificate diterbitkan otomatis. Ini membutuhkan paket **Cloudflare for SaaS** dan API token dengan izin
   `SSL and Certificates: Edit` (selain `Zone:DNS:Edit`).
3. Bila pendaftaran gagal, domain tetap disimpan dan routing tetap jalan; sertifikat bisa diterbitkan di zona
   Cloudflare milik user sendiri (proxy ke origin) atau dengan certificate manual.

Tanpa Cloudflare for SaaS, jangan arahkan domain langsung ke origin tanpa TLS.

### Webhook auto-deploy

Halaman project menampilkan *payload URL* (`https://<host>/webhooks/github`) dan secret per project. Daftarkan di
GitHub: **Settings → Webhooks → Add webhook**, content type `application/json`, event **push**. Setiap push ke
branch default memicu build baru. Signature `X-Hub-Signature-256` diverifikasi dengan secret project, dan endpoint
ini tidak memakai session.

### Upload ZIP

New project → **Upload ZIP** membuat project statis tanpa repository. Batas: 100 MB terkompresi, 500 MB setelah
diekstrak, 20.000 entri; symlink dan path traversal ditolak. Arsip harus berisi `index.html` di root atau di dalam
satu folder teratas. Project ZIP tidak bisa di-redeploy dari repo — buat project baru untuk versi berikutnya.

### Backup database berkala

Gunakan skrip yang memakai perintah `.backup` SQLite (aman saat server menulis):

```bash
cd /opt/only4.fun/backend
DB_PATH=/opt/only4.fun/backend/data/only4.fun.db ./scripts/backup-db.sh
```

Jadwalkan lewat cron, misal tiap 6 jam dan simpan 14 backup terakhir (default skrip):

```cron
0 */6 * * * /opt/only4.fun/backend/scripts/backup-db.sh >> /var/log/only4.fun/backup.log 2>&1
```

## Catatan

- Wildcard SSL (`*.only4.fun`) otomatis dari Cloudflare Universal SSL (gratis)
- Tidak perlu buat cert per subdomain — Cloudflare handle semua
- Subdomain baru (misal `myproject.only4.fun`) langsung bisa diakses tanpa konfigurasi tambahan
- Cloudflare cache bisa di-purge jika perlu: `curl -X POST "https://api.cloudflare.com/client/v4/zones/{zone_id}/purge_cache" -H "Authorization: Bearer {token}" -d '{"purge_everything":true}'`
- Untuk SSR, proses Node.js di-start otomatis saat pertama kali diakses
- Idle cleanup setiap 5 menit mematikan proses yang tidak diakses > 10 menit
