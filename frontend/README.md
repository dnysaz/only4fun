# Only4fun — Svelte landing page + launchpad

A [Svelte 5](https://svelte.dev) + [Vite](https://vite.dev) frontend for Only4fun,
converted from the original single-file HTML page into components with scoped styles.
Same layout, type and color system, plus a launchpad connected to the Go backend.

```bash
npm install
npm run dev       # dev server (proxies /api and /auth to the backend)
npm run build     # production build into dist/
npm run preview   # serve the build (same proxy)
```

## Running with the backend

Start the API first, then the frontend:

```bash
cd ../backend && go run ./cmd/server   # http://localhost:8080
cd ../frontend && npm run dev          # http://localhost:5173
```

- `/api` and `/auth` are proxied to `VITE_API_PROXY` (default `http://localhost:8080`),
  so session cookies stay same-origin and no backend CORS change is needed.
- For a deployed frontend, set `VITE_API_BASE` to the API origin and add that origin to
  the backend's `CORS_ORIGINS` (with credentials) — see `.env.example`.
- GitHub login goes through `GET /auth/github/login?next=/launchpad`; the backend sets
  the `session_id` cookie and redirects back.

## Structure

```
index.html            head, Google Fonts, meta/SEO tags
src/app.css           design tokens, base styles, shared utilities (.wrap, .btn, .mono, .dots4)
src/App.svelte        routes "/" (landing) and "/launchpad"
src/lib/
  api.js              fetch wrapper for the API (credentials: include)
  session.svelte.js   session store: /api/auth/me, sign out
  router.svelte.js    tiny History API router
  subdomain.js        shared subdomain rules (synced with the backend)
  Dots.svelte         four-dot logo mark
  Nav.svelte          header navigation (avatar + Dashboard when signed in)
  Hero.svelte         headline, subdomain checker, login hand-off
  DeployBand.svelte   red band + animated deploy log
  Statement.svelte    "Forget slow, complicated deploys" panel
  Illustration.svelte flat SVG artwork used by the statement panel
  Features.svelte     framework chips + feature grid
  Icon.svelte         stroke icon set used by the feature grid
  Cta.svelte          yellow call-to-action band
  Footer.svelte       footer
  Dashboard.svelte    sign-in gate, stats, create panel, project grid
  CreateProject.svelte  GitHub repo / ZIP tabs with live subdomain check
  ProjectCard.svelte    status, views, open / redeploy / delete
```

## Notes

- The hero checks names against `GET /api/check-subdomain` once signed in; guests get
  local validation plus a "sign in to claim it" hint.
- The launchpad polls `GET /api/projects` while any project is queued or building.
- Theming follows `prefers-color-scheme` and can be forced with
  `<html data-theme="dark">` or `data-theme="light"`.
- Contrast fixed vs. the original page: the step-badge labels and the CTA pill keep
  fixed ink colors, so they stay readable in dark mode.
