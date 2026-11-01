# Important — Production Notes

## Production Notes

- **Set `SESSION_SECRET` env var** (do not use default key `only4.fun-dev-insecure-default-key`). Without this, encryption of GitHub tokens and env vars is not secure.
- **Protect file `data/only4.fun.db`** — set file permission (chmod 600), regular backups, never expose to public.
- **Limit SSR process count** — `maxProcs = 50` in `ProcessManager`. Monitor and adjust if needed.
- **Monitor resource usage** (memory, CPU) for idle Node.js SSR processes. Idle cleanup every 5 minutes (disable idle processes > 10 minutes).
- **Consider rate limiting** for `/api/check-subdomain` endpoint (currently no specific rate limit).

## Architecture

- **Backend**: Go (Gin) + SQLite (`data/only4.fun.db`)
- **Frontend**: Next.js (App Router) — login, dashboard, project detail page, redeploy & delete
- **Build**: Direct host build (Node.js + npm on VPS)
- **Infrastructure**: VPS + Next.js :3000 + Node.js processes :4000-5000 (per project)

## Technical Architecture

```
Frontend        Backend API        Infrastructure
(Next.js)       (Go + Gin)       (Nginx + Cloudflare)
    │                  │                      │
    │                  │                      │
    └─► GitHub OAuth ─► Repo Clone ─► Build ─► Serve via Nginx
    │                  │                      │
    └─► Env Manager ─► Build Queue ─► Static Files ─► *.only4.fun
```

| Layer | Tech Stack | Purpose |
|-------|------------|----------|
| Frontend | Next.js (App Router) | Login, dashboard, project detail page, redeploy & delete |
| Backend | Go + Gin | API, GitHub OAuth, build orchestrator |
| Builder | Node.js + npm (host) | Clone repo, build static assets |
| Storage | Local disk / shared volume | Static files output |
| Web Server | Nginx | Reverse proxy per subdomain |
| DNS | Cloudflare API | Wildcard SSL (`*.only4.fun`) — automatic |

## Deploy Flow (User Journey)

1. User opens `only4.fun`, clicks "Login with GitHub"
2. Redirects to GitHub OAuth, returns to `only4.fun/dashboard`
3. User clicks "New Project", fills:
   - Git repo URL
   - Project name (becomes subdomain)
   - Environment variables (key:value, optional)
4. Backend clones repo → builds with Node.js/npm → saves static files
5. Backend adds DNS record to Cloudflare for `namaproject.only4.fun`
6. Nginx reload config, site goes live
7. User opens `namaproject.only4.fun`, sees their site

## API Endpoints (Backend Go Gin)

| Endpoint | Method | Purpose |
|----------|--------|----------|
| `/auth/github/login` | GET | Redirect to GitHub OAuth |
| `/auth/github/callback` | GET | Handle callback, create session |
| `/api/projects` | GET | List user projects |
| `/api/projects` | POST | Create new project (triggers deploy) |
| `/api/projects/:id` | GET | Project detail + status |
| `/api/projects/:id/deploy` | POST | Redeploy |
| `/api/projects/:id/env` | GET/POST/PATCH | Manage env vars |
| `/api/projects/:id` | DELETE | Delete project |
| `/api/auth/me` | GET | User info on login |

## Non-Functional

- **Build isolation**: Each build runs directly on the host (Node.js + npm)
- **Subdomain wildcard**: `*.only4.fun` in Cloudflare
- **Rate limit**: Prevent abuse of repo clone via GitHub OAuth
- **Health check**: Endpoint `/health` for monitoring

## Out of Scope (v1)

- Custom domain (user buys separately)
- CI/CD auto-deploy on GitHub push (webhook)
- Production-grade database (SQLite/boltdb for now)
- Multi-region / auto-scaling
- SSR / dynamic API routes (static only)

## Timeline (Roadmap)

| Phase | Task | Est. |
|-------|------|-------|
| 1 | Setup repo, Go Gin backend base, GitHub OAuth | 1-2 days |
| 2 | Frontend Next.js login + dashboard base | 1-2 days |
| 3 | Build pipeline (clone + build static) | 2-3 days |
| 4 | Subdomain + Cloudflare DNS + Nginx routing | 1-2 days |
| 5 | Env vars, redeploy, delete, polishing | 1-2 days |
| **Total** | **MVP ready to test** | **6-11 days** |

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| GitHub OAuth rate limit | Cache tokens, avoid excessive API calls |
| Build fails / timeout | Show clear error logs to user |
| Subdomain conflict | Validate unique project name in backend |
| Private repo clone without trust | Use GitHub token or deploy key |
| Nginx reload too often | Debounce reload, reload once per batch deploy |

### SSL & HTTPS

- Cloudflare Universal SSL (free) handles `only4.fun` and `*.only4.fun` automatically
- No per-subdomain certs needed — Cloudflare handles all
- Backend detects HTTPS via `X-Forwarded-Proto` header from Nginx
- Use Cloudflare **Full (strict)** SSL mode with Origin Certificate for best security

### Release Structure

```
data/sites/
├── <subdomain> -> .releases/<subdomain>/release-xxxx   (symlink)
│   ├── release-xxxx   (active release)
│   └── release-yyyy   (2 most recent backups)
```

Each deploy copies the build to a new release folder and atomically swaps the symlink. Three recent releases are kept; older ones are pruned automatically.

When a project is deleted, the symlink and `.releases/<subdomain>` folder are cleaned up.

### Custom Domain & SSL

- Custom domain set per project from the project page (`custom_domain` column)
- Create **CNAME** `domain-anda.com` → `<subdomain>.only4.fun` (or A record to VPS IP)
- Backend registers the hostname via Cloudflare API (`/zones/<zone>/custom_hostnames`) so edge certificates are auto-issued
- Without Cloudflare for SaaS, do not point domain directly to origin without TLS

### Webhook Auto-Deploy

- Project page shows *payload URL* (`https://<host>/webhooks/github`) and per-project secret
- Add webhook in GitHub: **Settings → Webhooks → Add webhook**, content type `application/json`, event **push**
- Every push to the default branch triggers a new build
- Signature `X-Hub-Signature-256` is verified with the project secret
- This endpoint does NOT use session-based auth

### ZIP Upload

- New project → **Upload ZIP** creates a static project without a repository
- Limits: 100 MB compressed, 500 MB extracted, 20,000 entries
- Archive must contain `index.html` at root or in the top-level folder
- ZIP cannot be redeployed from a repo — create a new project for the next version

### Periodic Database Backup

- Use the SQLite `.backup` script:
  ```bash
  cd /opt/only4.fun/backend
  DB_PATH=/opt/only4.fun/backend/data/only4.fun.db ./scripts/backup-db.sh
  ```
- Schedule via cron, e.g., every 6 hours
- Keep last 14 backups (default script)

## Security & Operations

- **Build isolation**: Builds run directly on the host (Node.js + npm available on PATH)
- **Subdomain wildcard**: `*.only4.fun` in Cloudflare Universal SSL (free)
- **No per-subdomain certs needed** — Cloudflare handles all
- **SSR process management**: Node.js processes per project (ports 4000-5000), auto-started on first access
- **Idle cleanup**: Every 5 minutes, kill SSR processes not used for >10 minutes

---
*Document version 1.0 — Production Guide*
