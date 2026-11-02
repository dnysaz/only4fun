# Only4fun

**Ship your site. Fast, Free And Simple.**

Only4fun (published as **only4.fun**) is an open-source, self-hostable hosting platform for static and
server-rendered websites. Point it at a GitHub repository — or upload a ZIP archive — and it builds the project on the
server, then serves the output on a subdomain such as `my-project.only4.fun` or on a domain you own.

[![CI](https://github.com/dnysaz/only4fun/actions/workflows/ci.yml/badge.svg)](https://github.com/dnysaz/only4fun/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-black.svg)](LICENSE)

> Only4fun is a project built for fun. It is free and open-source. See [`/terms`](frontend/app/terms/page.tsx) and [`/privacy`](frontend/app/privacy/page.tsx) in the app.

---

## Features

- **Deploy from GitHub** — pick a repository (public or private) and the platform clones, builds and publishes it.
- **Deploy without Git** — drop a ZIP archive of a static site.
- **Static and SSR** — static output is served as files; frameworks with a `start` script run as Node processes behind a reverse proxy.
- **Free subdomains** — every project gets `name.only4.fun`; project names are validated and filtered.
- **Custom domains** — attach your own domain with a CNAME record.
- **Zero-downtime deploys and rollback** — each deployment is a release folder; a new build is promoted only after it is ready, and the last three releases can be re-activated.
- **Auto-deploy webhooks** — a per-project GitHub webhook (HMAC-verified) rebuilds on every push to the default branch.
- **Environment variables** — encrypted at rest (AES-256-GCM), injected at build time, never returned to the browser.
- **Deployment logs** — streamed to the project page while a build runs.
- **Visit analytics** — aggregate page views per project per day, no user tracking.
- **GitHub profile extras** — bio, statistics, profile README and a contribution calendar on the Profile page.
- **Secure sessions** — GitHub OAuth, opaque session IDs stored only as SHA-256 hashes, 24-hour expiry, rate limits on write endpoints.
- **SEO ready** — per-page metadata, canonical URLs, `robots.txt`, `sitemap.xml`, Open Graph/Twitter images, JSON-LD and a web manifest.

## Architecture

```
        Browser
           │
           ▼
  Next.js app (frontend, :3000)
   ├── /dashboard, /docs, /profile …  (React UI)
   └── /api/* and /auth/*  ──rewrites──▶  Go API (backend, :8080)
                                              │
                        ┌─────────────────────┼──────────────────────┐
                        ▼                     ▼                      ▼
                 SQLite database       Build worker            Sites on disk
              users, sessions,      clone → build →          .releases/<subdomain>/
              projects, deploys     publish a release        <release> + proxy
                        │                     │                      │
                        └─────────── GitHub OAuth / API ─────────────┘
                                              │
                                     Cloudflare DNS (optional)
```

The Go API owns authentication, project state, the build queue, DNS records and the site proxy. The Next.js app is the
dashboard and the in-app documentation; it never talks to the database directly.

## Tech stack

| Layer | Technology |
|---|---|
| Frontend | Next.js 14 (App Router), React 18, TypeScript, CSS Modules |
| Backend | Go 1.23, Gin, `sqlx` |
| Database | SQLite |
| Build pipeline | Direct host builds (Node.js + npm) |
| Auth | GitHub OAuth 2.0 |
| DNS | Cloudflare API (optional) |

## Repository layout

```
.
├── backend/                 Go API server
│   ├── cmd/server/          entrypoint
│   ├── internal/
│   │   ├── builder/         clone, build, publish, worker queue
│   │   ├── config/          environment configuration
│   │   ├── database/        SQLite schema, stores, migrations
│   │   ├── dns/             Cloudflare DNS integration
│   │   ├── handler/         HTTP handlers (auth, projects, uploads, webhooks)
│   │   ├── id/              UUID v7 generator
│   │   ├── secure/          AES-256-GCM helpers, session hashing
│   │   └── server/          router, site proxy, process manager, rate limiting
│   └── scripts/             backup script
├── frontend/                Next.js dashboard, docs and marketing site
│   ├── app/                 routes (dashboard, projects, docs, help, login, …)
│   ├── components/          AppShell, Sidebar, AuthGate, ContributionGraph …
│   └── lib/                 API client, auth cache, site metadata, helpers
├── .github/workflows/       CI: Go vet/test, frontend typecheck/build
├── DEPLOY.md                production deployment guide (nginx, systemd, Cloudflare)
├── PRD.md                   product notes
└── LICENSE                  MIT
```

## Requirements

- **Go 1.23+**
- **Node.js 20+** and npm
- **Node.js + npm** (required for builds on the VPS host)
- A **GitHub OAuth app** (Client ID + secret) with callback `<api-host>/auth/github/callback`
- Optional: a **Cloudflare** API token if you want subdomain DNS created automatically

## Quick start

### 1. Backend

```bash
cd backend
cp .env.example .env        # then fill in the values
go run ./cmd/server
```

The API listens on `http://localhost:8080` (`GET /health` for a liveness check) and creates the SQLite database and
schema on first start.

Minimum values to set in `backend/.env`:

```ini
SESSION_SECRET=<openssl rand -hex 32>
GITHUB_CLIENT_ID=<your oauth app client id>
GITHUB_CLIENT_SECRET=<your oauth app client secret>
FRONTEND_URL=http://localhost:3000
```

### 2. Frontend

```bash
cd frontend
cp .env.local.example .env.local
npm install
npm run dev
```

Open `http://localhost:3000`, sign in with GitHub, and create your first project.

### 3. Production

`DEPLOY.md` covers running both services with systemd, the nginx reverse proxy, Cloudflare DNS, TLS and backups.

## Environment variables

### Backend (`backend/.env`)

| Variable | Default | Purpose |
|---|---|---|
| `SERVER_ADDR` | `:8080` | Listen address of the API |
| `FRONTEND_URL` | `http://localhost:3000` | Where users are redirected after login |
| `CORS_ORIGINS` | localhost origins | Comma-separated allowed origins |
| `DATABASE_PATH` | `./data/only4.fun.db` | SQLite file |
| `SESSION_SECRET` | insecure dev default | **Required in production.** Encrypts tokens, hashes sessions |
| `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` | – | GitHub OAuth app credentials |
| `GITHUB_REDIRECT_URL` | `http://localhost:8080/auth/github/callback` | OAuth callback URL |
| `CLOUDFLARE_API_KEY` / `CLOUDFLARE_ZONE_ID` / `CLOUDFLARE_DOMAIN` | – | Automatic subdomain DNS (optional) |
| `CLOUDFLARE_TARGET` | – | VPS IP (A record) or hostname (CNAME) for subdomains |
| `CLOUDFLARE_PROXY` | `true` | Send traffic through the Cloudflare proxy |
| `CACHE_DIR` | `/tmp/only4.fun/cache` | Temporary clone/build workspace |
| `SITES_DIR` | `/var/lib/only4.fun/sites` | Published releases served by the proxy |
| `BUILD_TIMEOUT_SECONDS` | `600` | Timeout per build |

### Frontend (`frontend/.env.local`)

| Variable | Default | Purpose |
|---|---|---|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | Base URL of the Go API |
| `NEXT_PUBLIC_SITE_URL` | `https://only4.fun` | Public origin used for canonical URLs, sitemap and Open Graph |

GitHub credentials belong to the backend only — never place a client secret in the frontend, because everything bundled
there is visible to visitors.

## API overview

All `/api/*` routes require the session cookie. Full reference: **`/docs/api`** in the running app.

| Method | Endpoint | Purpose |
|---|---|---|
| `GET` | `/health` | Liveness check |
| `GET` | `/auth/github/login?next=/path` | Start OAuth; remembers the target page |
| `GET` | `/auth/github/callback` | Finish OAuth, set the session cookie |
| `GET` | `/api/auth/me` | Current user |
| `POST` | `/api/auth/logout` | Delete the session |
| `GET` / `POST` | `/api/projects` | List / create projects |
| `POST` | `/api/projects/upload` | Create a static project from a ZIP |
| `GET` / `DELETE` | `/api/projects/:id` | Project detail / delete |
| `POST` | `/api/projects/:id/deploy` | Queue a rebuild |
| `GET` | `/api/projects/:id/deployments` | Deployment history with logs |
| `POST` | `/api/projects/:id/deployments/:deploymentID/rollback` | Re-activate an older release |
| `GET` / `PUT` | `/api/projects/:id/env` | Read env keys / update env vars |
| `PUT` | `/api/projects/:id/domain` | Set or clear a custom domain |
| `GET` | `/api/projects/:id/webhook` | Webhook URL and secret |
| `GET` | `/api/github/repos?page=N` | Repositories available to deploy |
| `GET` | `/api/github/profile` | Bio, README and contribution calendar |
| `POST` | `/webhooks/github` | GitHub push webhook (HMAC verified, no session) |

## Development

```bash
# Backend
cd backend
go test ./...          # unit tests
go vet ./...           # static checks
gofmt -l .             # formatting check (CI fails on output)

# Frontend
cd frontend
npx tsc --noEmit       # type check
npm run build          # production build
```

CI runs exactly these steps on every push and pull request.

## Security

- Sessions are random 256-bit identifiers; the database stores only their SHA-256 hash.
- GitHub tokens and environment values are encrypted with AES-256-GCM using `SESSION_SECRET`.
- Every project endpoint verifies ownership.
- Builds run directly on the host (Node.js + npm).
- ZIP uploads are limited and reject path traversal and symlinks.
- GitHub webhooks are verified with an HMAC SHA-256 signature.

**Never commit `.env` files, the SQLite database, or any credential.** They are covered by `.gitignore`; keep them that
way. If you find a security issue, please report it privately instead of opening a public issue.

## Companion project: Moogo

Only4fun has no built-in database. For projects that need one, the recommended companion is
**[Moogo](https://moogo.dev)** — SQLite over HTTP, also free and open source, made by the same author.
See **`/docs/database`** in the app for the full integration guide.

## Contributing

1. Fork the repository and create a branch: `git checkout -b feat/my-change`.
2. Keep the code formatted (`gofmt`) and the type check clean (`npx tsc --noEmit`).
3. Add or update tests where behaviour changes.
4. Open a pull request describing the problem and the change.

Bug reports and feature ideas are welcome through the issue tracker.

## License

[MIT](LICENSE) © 2026 only4.fun contributors.

## Credits

Built with Go, Gin, SQLite, Next.js and React. DNS automation through Cloudflare. Database companion: Moogo.
