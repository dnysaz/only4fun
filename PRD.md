# PRD — only4.fun

Static Site Hosting / Serverless (simple, Vercel-like) with GitHub login, deploy from Git repo, env vars, and automatic subdomain.

---

## 1. Objective

Build a simple platform so users can publish static/serverless sites from a Git repo, log in via GitHub, and instantly get a `*.only4.fun` subdomain.

---

## 2. Users (Personas)

| Type | Needs |
|------|-------|
| User | GitHub login, paste Git repo URL, set env vars, get subdomain, view deploy status |
| Admin (optional) | View project list, manage quotas, logout users |

---

## 3. Core Features (MVP)

### 3.1 Authentication
- Login only via GitHub OAuth
- Session stored in backend (cookie/JWT)
- Frontend shows "Login with GitHub" button

### 3.2 Deploy from Git Repo
- User pastes Git repo URL (public/private)
- Backend clones repo
- Detect framework/build command (optional default: `npm run build` with output `dist` or `out`)
- Build directly on host (Node.js + npm)
- Save build output to shared storage

### 3.3 Environment Variables
- User inputs key:value pairs in UI before/during deploy
- ENV injected into build process
- Show list of env vars set per project

### 3.4 Automatic Subdomain
- Each project gets subdomain: `projectname.only4.fun`
- DNS added automatically to Cloudflare (via API)
- Nginx reverse proxy configured for subdomain routing

### 3.5 Dashboard
- List of deployed projects
- Deploy status (pending/building/deployed/failed)
- Buttons: redeploy, delete project, edit env vars

---

## 4. Technical Architecture

```
Frontend        Backend API         Infrastructure
(Next.js)       (Go + Gin)          (Nginx + Cloudflare)
    │                  │                      │
    │                  │                      │
    └─► GitHub OAuth ─► Repo Clone ─► Build ─► Serve via Nginx
    │                  │                      │
    └─► Env Manager ─► Build Queue ─► Static Files ─► *.only4.fun
```

| Layer | Tech Stack | Purpose |
|-------|------------|---------|
| Frontend | Next.js (App Router) | Login UI, dashboard, deploy form, env input |
| Backend | Go + Gin | API, GitHub OAuth, build orchestrator |
| Builder | Node.js + npm | Clone repo, build static assets |
| Storage | Local disk / shared volume | Static files output |
| Web Server | Nginx | Reverse proxy per subdomain |
| DNS | Cloudflare API | Auto-create A/CNAME records for `*.only4.fun` |

---

## 5. Deploy Flow (User Journey)

1. User opens `only4.fun`, clicks "Login with GitHub"
2. Redirects to GitHub OAuth, returns to `only4.fun/dashboard`
3. User clicks "New Project", fills:
   - Git repo URL
   - Project name (becomes subdomain)
   - Environment variables (key:value, optional)
4. Backend clones repo → builds with Node.js/npm → saves static files
5. Backend adds DNS record to Cloudflare for `namaproject.only4.fun`
6. Nginx reloads config, site goes live
7. User opens `namaproject.only4.fun`, sees their site

---

## 6. API Endpoints (Backend Go Gin)

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/auth/github/login` | GET | Redirect to GitHub OAuth |
| `/auth/github/callback` | GET | Handle callback, create session |
| `/api/projects` | GET | List user projects |
| `/api/projects` | POST | Create new project (triggers deploy) |
| `/api/projects/:id` | GET | Project detail + status |
| `/api/projects/:id/deploy` | POST | Redeploy |
| `/api/projects/:id/env` | GET/POST/PATCH | Manage env vars |
| `/api/projects/:id` | DELETE | Delete project |
| `/api/auth/me` | GET | Logged-in user info |

---

## 7. Non-Functional Requirements

- **Build isolation**: Each build runs directly on the host (Node.js + npm)
- **Subdomain wildcard**: `*.only4.fun` in Cloudflare
- **Rate limit**: Prevent abuse of repeated repo cloning
- **Health check**: `/health` endpoint for monitoring

---

## 8. Out of Scope (v1)

- Custom domain (user provides their own)
- CI/CD auto-deploy on GitHub push (webhook)
- Production-grade database (SQLite/boltdb for now)
- Multi-region / auto-scaling
- Server-side rendering / dynamic API routes (static only)

---

## 9. Timeline (Initial Plan)

| Phase | Task | Est. |
|-------|------|------|
| 1 | Setup repo, Go Gin backend base, GitHub OAuth | 1-2 days |
| 2 | Frontend Next.js login + dashboard base | 1-2 days |
| 3 | Build pipeline (clone + build static) | 2-3 days |
| 4 | Subdomain + Cloudflare DNS + Nginx routing | 1-2 days |
| 5 | Env vars, redeploy, delete, polishing | 1-2 days |
| **Total** | **MVP ready to test** | **6-11 days** |

---

## 10. Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| GitHub OAuth rate limit | Cache tokens, don't call API excessively |
| Build fails / timeout | Limit build time, show error logs to user |
| Subdomain conflict | Validate unique project name in backend |
| Private repo clone without trust | Use GitHub token or deploy key |
| Nginx reload too often | Debounce reload, reload once per batch deploy |

---

*Created as shared reference while building only4.fun. This document may be updated as development progresses.*