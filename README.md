<img src="assets/logo.svg" alt="Only4fun" width="96" />

# Only4fun

**Ship any static sites in seconds.**

Only4fun is a small, self-hosted deployment platform for fun side projects. Paste a
public GitHub repository or drop a ZIP and the site goes live at a
`yoursite.only4.fun` address. There are two ways to deploy:

- **Instant Launch** (`/instant`) — no account needed. Paste a public GitHub URL or
  upload a ZIP, accept the terms, and get an eight-character random address such as
  `asj3jkha.only4.fun`. The address shown in the preview is exactly the one that gets
  deployed. Instant deploy accepts static HTML, CSS and JavaScript only; framework
  projects are rejected with a prompt to sign in.
- **Deploy new** (`/deploy`) — for signed-in GitHub users. Search and pick one of your
  GitHub repositories (public or private) or upload a ZIP, optionally add environment
  variables at creation, then publish at a chosen subdomain.
- **Launchpad** (`/launchpad`) — for signed-in GitHub users. Full framework builds,
  a project list with traffic stats, environment variables, custom domains,
  deployment history and rollback.

Instant projects are owned by a reserved `anonymous` account and never appear in
anyone's launchpad.

## How it works

1. **Add files** — connect a repository or upload a ZIP.
2. **Deploy** — the backend clones or extracts the source and builds it.
3. **SSL** — Cloudflare issues the certificate and proxies the subdomain.
4. **Launch** — the site is published at `<name>.only4.fun`, ready to share.

## Features

- Flat, colorful Svelte 5 interface with a floating pill navbar and a dark footer.
- Eight-character random subdomains for anonymous deploys, with reservation tokens
  so the previewed address is the delivered address.
- Framework builds for Next.js, Astro, Vue, React, Svelte, Nuxt, Angular, Solid,
  Vite and plain HTML/JS, using the official `simple-icons` logos.
- `npm`, `pnpm` and `yarn` detected from the lockfile; `npm run build` by default.
- Immutable release directories with symlink switching; the last three releases
  are kept and can be rolled back.
- Per-project environment variables, custom domains, visit counters and a 14-day
  traffic chart.
- Real info pages: Docs, Status, Support, Terms and Privacy.
- Exact-match blocked-subdomain list (366 words) enforced in both the frontend and
  the backend.

## Tech stack

| Layer | Technology |
| --- | --- |
| Frontend | Svelte 5, Vite 8, plain CSS, `simple-icons` |
| Backend | Go 1.23, Gin, `sqlx` |
| Database | SQLite (`mattn/go-sqlite3`), single local file |
| Build | Host-side `git clone` / ZIP extraction, `npm` / `pnpm` / `yarn` |
| Publishing | Release directories plus symlinks under `data/sites` |
| DNS / TLS | Optional Cloudflare API integration |

## Repository layout

```
.
├── assets/
│   └── logo.svg                  Four-dot Only4fun logo used in this README
├── backend/
│   ├── .env.example              Backend configuration template (no secrets)
│   ├── cmd/server/main.go        Server entry point: config, DB, stores, router
│   ├── internal/
│   │   ├── builder/              Clone/extract, build, publish, release + rollback
│   │   ├── config/               Environment variable loading
│   │   ├── database/             Schema, migrations, users/sessions/projects/deployments
│   │   ├── dns/                  Cloudflare DNS record management
│   │   ├── handler/              HTTP handlers: auth, projects, uploads, instant, webhooks
│   │   ├── id/                   UUID v7 generation
│   │   ├── secure/               Session IDs, AES-GCM encryption, CSRF helpers
│   │   └── server/               Router, middleware, rate limits, static site serving
│   ├── scripts/backup-db.sh      SQLite backup helper
│   └── data/                     Local runtime data (git-kept placeholder only)
├── frontend/
│   ├── public/favicon.svg        Four-dot favicon
│   ├── public/og-image.png       Social sharing card (1200 × 630)
│   ├── public/robots.txt         Crawler rules and sitemap link
│   ├── public/sitemap.xml        Indexable public routes
│   ├── src/
│   │   ├── app.css               Design tokens, base styles, shared utilities
│   │   ├── App.svelte            Route switchboard
│   │   ├── main.js               App bootstrap
│   │   └── lib/
│   │       ├── Dots.svelte       Four-dot logo mark
│   │       ├── Nav.svelte        Floating navigation bar
│   │       ├── Hero.svelte       Landing page headline and subdomain claim
│   │       ├── DeployBand.svelte Red deploy banner with animated log
│   │       ├── Statement.svelte  How-it-works panel
│   │       ├── Features.svelte   Framework marquee and feature grid
│   │       ├── Cta.svelte        Yellow call-to-action band
│   │       ├── Footer.svelte     Dark footer with links and watermark
│   │       ├── Workspace.svelte  Signed-in launchpad shell
│   │       ├── Dashboard.svelte  Launchpad layout and stats
│   │       ├── CreateProject.svelte  Repository / ZIP deploy form
│   │       ├── ProjectDetail.svelte  Single project page
│   │       ├── InstantDeploy.svelte  No-login deploy form
│   │       ├── InfoPage.svelte   Shared Docs/Status/Support/Terms/Privacy layout
│   │       ├── router.svelte.js  Tiny History API router
│   │       ├── seo.js            Route-aware titles, descriptions, canonical tags
│   │       ├── session.svelte.js Session store
│   │       ├── workspace.svelte.js Project store and polling
│   │       ├── api.js            Fetch wrapper with credentials
│   │       ├── subdomain.js      Client-side subdomain rules
│   │       ├── check_word.ts     Blocked-word list (mirrored in the backend)
│   │       ├── frameworks.ts     Official simple-icons metadata
│   │       ├── fileTree.js       Repository tree preview helpers
│   │       ├── zipTree.js        ZIP tree preview helpers
│   │       └── infoPages.js      Info page content
│   ├── .env.example              Frontend configuration template
│   ├── index.html                HTML shell, fonts, meta tags
│   └── vite.config.js            Dev proxy for /api and /auth
├── LICENSE                       MIT license
└── README.md                     This file
```

## Getting started

Prerequisites: Go 1.23+, Node.js with npm, and a local loopback for the dev login.
SQLite is embedded in the Go binary, so no separate database server is required.

### 1. Backend

```bash
cd backend
cp .env.example .env
# Fill in SESSION_SECRET, GitHub OAuth, and Cloudflare values as needed.
go run ./cmd/server
```

The API listens on `http://localhost:8080` by default. Migrations run on startup and
the SQLite file is created at `data/only4.fun.db`. The `anonymous` system account is
created automatically when the instant handlers are initialized.

For local development only, `DEV_LOGIN=true` enables `GET /auth/dev-login`, which
creates a demo session and three sample projects. The endpoint rejects requests
that do not come from loopback. **Never enable it in production.**

### 2. Frontend

```bash
cd frontend
npm install
npm run dev       # http://localhost:5173
```

`/api` and `/auth` are proxied to `VITE_API_PROXY` (default
`http://localhost:8080`), so session cookies stay same-origin. For a separately
hosted frontend, set `VITE_API_BASE` to the API origin and add that origin to the
backend's `CORS_ORIGINS`.

```bash
npm run build     # production build into dist/
npm run preview   # serve the production build locally
```

### 3. Tests

```bash
cd backend && go test ./...
cd frontend && npm run build
```

## Configuration

Backend variables (see `backend/.env.example`):

| Variable | Default | Purpose |
| --- | --- | --- |
| `SERVER_ADDR` | `:8080` | API listen address |
| `FRONTEND_URL` | `http://localhost:5173` | Post-login redirect target |
| `CORS_ORIGINS` | localhost origins | Comma-separated allowed origins |
| `DEV_LOGIN` | `false` | Enable loopback-only demo login |
| `DATABASE_PATH` | `data/only4.fun.db` | SQLite file path |
| `SESSION_SECRET` | empty | Encrypts tokens and hashes sessions; required in production |
| `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET` | empty | GitHub OAuth app credentials |
| `GITHUB_REDIRECT_URL` | `http://localhost:8080/auth/github/callback` | OAuth callback |
| `CLOUDFLARE_API_KEY`, `CLOUDFLARE_ZONE_ID` | empty | Enable Cloudflare automation when both are set; token needs Zone:DNS:Edit and Zone:SSL and Certificates:Edit |
| `CLOUDFLARE_DOMAIN` | `only4.fun` | Zone used for project subdomains; Cloudflare for SaaS must be enabled for custom-hostname SSL |
| `CLOUDFLARE_TARGET` | empty | A record IP or CNAME target |
| `CLOUDFLARE_PROXY` | `true` | Route traffic through Cloudflare (TLS at the edge) |
| `CACHE_DIR` | `/tmp/only4.fun/cache` | Temporary clone and build workspace |
| `SITES_DIR` | `./data/sites` | Published sites and release directories |
| `BUILD_TIMEOUT_SECONDS` | `600` | Per-build timeout |

Frontend variables (see `frontend/.env.example`):

| Variable | Purpose |
| --- | --- |
| `VITE_API_BASE` | Absolute API origin; leave empty to use the dev proxy |
| `VITE_API_PROXY` | Dev/preview proxy target, normally `http://localhost:8080` |
| `VITE_CUSTOM_DOMAIN_CNAME` | CNAME target shown in the custom-domain hint, default `customers.only4.fun` |

## Routes and API

Frontend routes:

| Route | Page |
| --- | --- |
| `/` | Landing page |
| `/instant` | No-login static deploy |
| `/launchpad` | Your projects (sign-in required) |
| `/deploy` | Deploy new (sign-in required) |
| `/project/:id` | Project detail, settings and history |
| `/docs`, `/status`, `/support`, `/terms`, `/privacy` | Info pages |

Backend endpoints:

| Group | Endpoints |
| --- | --- |
| Public | `GET /health`, `GET /metrics`, `POST /webhooks/github` |
| Instant | `POST /api/instant/address`, `POST /api/instant/repo-tree`, `POST /api/instant/repo`, `POST /api/instant/zip` |
| Auth | `GET /auth/github/login`, `GET /auth/github/callback`, `GET /api/auth/me`, `POST /api/auth/logout` |
| Projects | `GET/POST /api/projects`, `GET/DELETE /api/projects/:id`, `POST /api/projects/upload`, `POST /api/projects/:id/deploy`, `GET/PUT /api/projects/:id/domain`, `GET /api/projects/:id/domain/status`, `GET /api/projects/:id/env`, `GET /api/projects/:id/repo`, `GET /api/projects/:id/webhook` |
| Deployments | `GET /api/projects/:id/deployments`, `POST /api/projects/:id/deployments/:deploymentID/rollback`, `DELETE /api/projects/:id/deployments` |
| GitHub | `GET /api/github/repos`, `GET /api/github/repos/:owner/:repo/contents`, `GET /api/github/repos/:owner/:repo/readme`, `GET /api/github/profile` |
| Dev only | `GET /auth/dev-login` (loopback only, `DEV_LOGIN=true`) |

## Deployment notes

- Run builds only from repositories you trust: install and build scripts execute on
  the host with the server process user. Builds are always local (`build mode: local`).
- Keep `DEV_LOGIN` disabled in production, set a strong random `SESSION_SECRET`, and
  restrict `CORS_ORIGINS` to your real frontend origins.
- `backend/data/` holds the SQLite database and published sites. It is runtime state,
  not source code; only a `.gitkeep` placeholder is tracked.
- Cloudflare DNS automation stays off when `CLOUDFLARE_API_KEY` and
  `CLOUDFLARE_ZONE_ID` are empty. Without it, no DNS records are created.
- Custom-domain SSL needs Cloudflare for SaaS enabled on the zone, an active
  Fallback Origin, and a token with `Zone > SSL and Certificates > Edit`.
  Create a proxied CNAME target (`customers.only4.fun` → `fallback.only4.fun`)
  and ask users to point their domain's CNAME to it; an A record to the origin
  IP is not supported without the apex-proxying entitlement. The UI shows a
  warning when custom-hostname registration fails; domain storage and
  Host-based serving still work when DNS points to the server.
- Public endpoints are protected by a trusted-origin guard and per-IP rate limits.
  Uploads are size-limited and checked for path traversal and symlinks; GitHub URLs
  are validated before cloning, and SQL queries are parameterized.

## License

[MIT](LICENSE) © 2026 dnysaz
