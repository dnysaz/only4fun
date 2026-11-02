# Deployment Guide — VPS + Cloudflare

## Architecture

```
User Browser
     │
     ▼
Cloudflare (SSL termination + WAF + DDoS protection + proxy)
     │  HTTPS (edge cert: *.only4.fun)
     ▼
VPS (Go backend :8080 + Next.js :3000)
     │  HTTP (internal, not exposed to public)
     ▼
data/sites/<subdomain>/  (static files)
     │
     ▼ (if SSR)
Node.js process :4000-5000 (per project)
```

## SSL / Subdomain

**No need to configure SSL per subdomain.** Cloudflare Universal SSL automatically covers:
- `only4.fun`
- `*.only4.fun` (all subdomains)

Just set a wildcard DNS record in Cloudflare:
```
Type: A
Name: *
Content: <VPS_IP>
Proxy: Proxied (orange cloud) ← REQUIRED for SSL to be active
TTL: Auto
```

And one record for the root:
```
Type: A
Name: @ (or only4.fun)
Content: <VPS_IP>
Proxy: Proxied
```

### Cloudflare SSL Mode

Use **Full (strict)**:
1. Cloudflare → SSL/TLS → Overview → Full (strict)
2. Generate Origin Certificate: SSL/TLS → Origin Server → Create Certificate
3. Install cert on VPS (nginx/caddy) or Go backend

Or use **Flexible** (easier, but less secure):
- Cloudflare handles SSL to user, VPS serves HTTP
- No cert needed on VPS

## deploy.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

# ─── Configuration ──────────────────────────────────────────────────────────────
APP_DIR="/opt/only4.fun"
BACKEND_DIR="$APP_DIR/backend"
FRONTEND_DIR="$APP_DIR/frontend"
LOG_DIR="/var/log/only4.fun"
SERVICE_BACKEND="only4fun-backend"
SERVICE_FRONTEND="only4fun-frontend"

# ─── Output colors ─────────────────────────────────────────────────────────────
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
# IMPORTANT: set NEXT_PUBLIC_API_URL for production
NEXT_PUBLIC_API_URL=https://only4.fun npm run build || err "frontend build failed"
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
Environment="GITHUB_CLIENT_ID=your-client-id"
Environment="GITHUB_CLIENT_SECRET=your-client-secret"
Environment="GITHUB_REDIRECT_URL=https://only4.fun/auth/github/callback"
Environment="FRONTEND_URL=https://only4.fun"
Environment="CLOUDFLARE_API_KEY=your-api-token"
Environment="CLOUDFLARE_ZONE_ID=your-zone-id"
Environment="CLOUDFLARE_DOMAIN=only4.fun"
Environment="CLOUDFLARE_TARGET=your-vps-ip"
Environment="CLOUDFLARE_PROXY=true"
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
Environment="NEXT_PUBLIC_API_URL=https://only4.fun"
ExecStart=/usr/bin/npm start
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## Reverse Proxy (Nginx)

**IMPORTANT:** The Nginx config must route subdomain requests (e.g. `ytg.only4.fun`) to the backend
(port 8080) so the backend's SiteHandler can serve the deployed site. Main domain requests go to the
frontend (port 3000).

```nginx
# /etc/nginx/sites-available/only4.fun

# Map host to backend port: main domain -> frontend (3000), subdomains -> backend (8080)
map $host $backend_port {
    only4.fun       3000;
    www.only4.fun   3000;
    default         8080;
}

server {
    listen 80;
    server_name only4.fun *.only4.fun;

    # Redirect HTTP to HTTPS
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name only4.fun *.only4.fun;

    # Cloudflare Origin Certificate
    ssl_certificate /etc/ssl/certs/cloudflare-origin.pem;
    ssl_certificate_key /etc/ssl/private/cloudflare-origin.key;

    # All requests: route to frontend (3000) or backend (8080) based on host
    location / {
        proxy_pass http://127.0.0.1:$backend_port;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # Backend API (always backend)
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Auth routes (always backend)
    location /auth/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Health check (always backend)
    location /health {
        proxy_pass http://127.0.0.1:8080;
    }
}
```

**IMPORTANT:** The `map` block must be placed in the `http` context (outside the `server` block), typically in `/etc/nginx/nginx.conf` or in a file included from there. If you put it inside the server block, Nginx will fail to start.

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

## Security & Operations

### Build Isolation

Builds run directly on the host (Node.js + npm). Make sure Node.js and npm are installed and available on the PATH. Build output is published to `data/sites/<subdomain>/`.

### HTTPS & Cookies

Because TLS terminates at Cloudflare/nginx, the backend detects HTTPS from the proxy header
(`X-Forwarded-Proto: https` or `CF-Visitor: {"scheme":"https"}`) and only then sends the
`Secure` cookie. Make sure Nginx forwards that header (the config above already does):

```nginx
proxy_set_header X-Forwarded-Proto $scheme;
```

Use Cloudflare SSL mode **Full (strict)** with Origin Certificate. **Flexible** mode keeps the
Cloudflare → VPS hop as plain HTTP, so session cookies and GitHub tokens can be intercepted on the network.

### Release Structure & Atomic Deploy

```
data/sites/
├── <subdomain> -> .releases/<subdomain>/release-xxxx   (symlink)
└── .releases/<subdomain>/
    ├── release-xxxx   (active release)
    └── release-yyyy   (2 most recent backups)
```

Each deploy copies the build to a new release folder and atomically swaps the symlink, so there
is no moment when the subdomain loses its active folder. Three recent releases are kept; the rest
are pruned automatically. When a project is deleted, the symlink and `.releases/<subdomain>` folder
are cleaned up as well.

### Custom Domain & SSL

Custom domains are configured per project from the project page (`custom_domain` column). After saving:
1. Create a **CNAME** `yourdomain.com` → `<subdomain>.only4.fun` (or A record to VPS IP).
2. Backend registers that hostname via the Cloudflare API (`/zones/<zone>/custom_hostnames`) so an
   edge certificate is auto-issued. This requires the **Cloudflare for SaaS** add-on and an API token
   with `SSL and Certificates: Edit` permission (in addition to `Zone:DNS:Edit`).
3. If registration fails, the domain is still saved and routing still works; the certificate can be
   issued in the user's own Cloudflare zone (proxy to origin) or with a manual certificate.

Without Cloudflare for SaaS, do not point a domain directly to the origin without TLS.

### Webhook Auto-Deploy

The project page shows a *payload URL* (`https://<host>/webhooks/github`) and a per-project secret.
Register it in GitHub: **Settings → Webhooks → Add webhook**, content type `application/json`, event **push**.
Every push to the default branch triggers a new build. The `X-Hub-Signature-256` signature is verified
with the project secret, and this endpoint does not use sessions.

### ZIP Upload

New project → **Upload ZIP** creates a static project without a repository. Limits: 100 MB compressed,
500 MB extracted, 20,000 entries; symlinks and path traversal are rejected. The archive must contain
`index.html` at the root or inside the top-level folder. ZIP projects cannot be redeployed from a repo —
create a new project for the next version.

### Periodic Database Backup

Use the script that uses the SQLite `.backup` command (safe while the server is writing):

```bash
cd /opt/only4.fun/backend
DB_PATH=/opt/only4.fun/backend/data/only4.fun.db ./scripts/backup-db.sh
```

Schedule via cron, e.g. every 6 hours, keeping the last 14 backups (script default):

```cron
0 */6 * * * /opt/only4.fun/backend/scripts/backup-db.sh >> /var/log/only4.fun/backup.log 2>&1
```

## Notes

- Wildcard SSL (`*.only4.fun`) is automatic from Cloudflare Universal SSL (free)
- No per-subdomain cert needed — Cloudflare handles all
- New subdomains (e.g. `myproject.only4.fun`) work immediately without extra configuration
- Cloudflare cache can be purged if needed: `curl -X POST "https://api.cloudflare.com/client/v4/zones/{zone_id}/purge_cache" -H "Authorization: Bearer {token}" -d '{"purge_everything":true}'`
- For SSR, the Node.js process starts automatically on first access
- Idle cleanup every 5 minutes kills processes not accessed for > 10 minutes