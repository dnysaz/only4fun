# Audit Fix Plan

Catatan progres perbaikan hasil audit BE + FE tanggal 2026-11-27.
**Aturan:** satu temuan dikerjakan satu per satu, status di-update setelah verifikasi, lalu commit.

## Status

- `Pending` — belum dikerjakan
- `Doing` — sedang dikerjakan
- `Done` — sudah diperbaiki & diverifikasi
- `Skipped` — sengaja tidak dilakukan (alasan dicatat)

Verifikasi standar: `cd backend && go test ./... -count=1`, `go vet ./...`, `cd frontend && npm run build`, `git diff --check`, secret scan.

## Critical

| ID | Temuan | File utama | Status | Catatan |
|---|---|---|---|---|
| C1 | Build/SSR mewarisi semua secret server lewat `os.Environ()` | `backend/internal/builder/build.go`, `backend/internal/server/process_manager.go` | Done | `internal/execenv` hanya mengirim `PATH`, `HOME`, `TMPDIR`, cache npm, dan `CI`; dipakai untuk git/build/SSR. Isolasi container/VM tetap rekomendasi operasional. |
| C2 | ZIP authenticated bisa menyimpan `server.js` dan memicu SSR/RCE | `backend/internal/handler/upload.go`, `backend/internal/server/sites.go` | Done | Upload menolak semua kandidat server entry; SiteHandler tidak pernah menjalankan SSR untuk project `zip://`; `server.js` repo juga hanya SSR bila ada script `start`. |

## High

| ID | Temuan | File utama | Status | Catatan |
|---|---|---|---|---|
| H1 | Token GitHub revoked hanya dicek di `/auth/me`, bukan middleware | `backend/internal/handler/auth_middleware.go`, `auth.go` | Pending | Validasi terpusat + cache singkat. |
| H2 | Origin guard mempercayai `Host` request | `backend/internal/server/csrf.go` | Pending | Hapus shortcut Host. |
| H3 | Trusted proxy / IP mismatch merusak rate limit & cookie Secure | `backend/internal/server/router.go`, `ratelimit.go` | Pending | Konfigurasi trusted proxy. |
| H4 | `http.Server` tanpa timeout | `backend/internal/server/server.go` | Pending | Tambah read/idle/write timeout. |
| H5 | Fallback encryption key publik saat `SESSION_SECRET` kosong | `backend/internal/secure/secure.go` | Pending | Fail startup di production. |
| H6 | Data project lama tampil saat pindah ke project lain yang gagal dimuat | `frontend/src/lib/ProjectDetail.svelte` | Pending | Clear state saat `id` berubah. |
| H7 | Guard availability:/submit belum menunggu hasil check sukses | `frontend/src/lib/CreateProject.svelte`, `Hero.svelte` | Pending | Wajib `available === true`. |
| H8 | Repo URL arbitrary bisa lolos ke submit | `frontend/src/lib/CreateProject.svelte` | Pending | Validasi hostname GitHub. |
| H9 | Instant deploy bisa beda alamat dari preview | `frontend/src/lib/InstantDeploy.svelte` | Pending | Wajib token+subdomain preview. |
| H10 | Revalidasi session hanya saat `visibilitychange` | `frontend/src/App.svelte`, `session.svelte.js` | Pending | Revalidate saat route change + idle timer. |
| H11 | Poll tetap jalan setelah session expired | `frontend/src/lib/ProjectDetail.svelte` | Pending | Gate poll dengan `session.user`. |

## Medium

| ID | Temuan | File utama | Status | Catatan |
|---|---|---|---|---|
| M1 | Custom domain tidak dihapus dari Cloudflare saat project delete | `backend/internal/handler/project.go` | Pending | Panggil `DeleteCustomHostname`. |
| M2 | Unique index custom_domain case-sensitive | `backend/internal/database/migrate.go` | Pending | Index `lower(custom_domain)`. |
| M3 | SSR startup menahan global mutex sampai 10 detik | `backend/internal/server/process_manager.go` | Pending | Lock per subdomain. |
| M4 | Env key bisa override `PATH`, `LD_PRELOAD`, `NODE_OPTIONS` | `backend/internal/handler/project.go`, `builder/build.go` | Pending | Deny-list + namespace. |
| M5 | Nil panic saat housekeeping release | `backend/internal/builder/build.go` | Pending | Cek error `Info()`. |
| M6 | Metrics counters data race | `backend/internal/server/metrics.go`, `middleware.go` | Pending | Atomic/mutex. |
| M7 | Foreign keys tidak aktif untuk DB baru | `backend/internal/database/init.go` | Pending | `PRAGMA foreign_keys=ON`. |
| M8 | GitHub 403 dianggap token revoked | `backend/internal/handler/auth.go` | Pending | Hanya 401 yang fatal. |
| M9 | Clone publik tanpa kuota ukuran repo | `backend/internal/handler/instant.go` | Pending | Cek size + disk quota. |
| M10 | ZIP index resolution terlalu kaku | `backend/internal/handler/upload.go` | Pending | Cari root berisi `index.html`. |
| M11 | Webhook secret dibuat sebelum HMAC terverifikasi | `backend/internal/handler/webhook.go` | Pending | Hanya baca secret. |
| M12 | Race session/workspace cleanup | `frontend/src/lib/session.svelte.js`, `workspace.svelte.js` | Pending | Generation counter + invalidate terpusat. |
| M13 | Deploy history dibatasi 6 padahal backend kirim 20 | `frontend/src/lib/ProjectDetail.svelte` | Pending | Hilangkan cap / pagination nyata. |
| M14 | Key grafik memakai lokal vs data UTC | `frontend/src/lib/ProjectDetail.svelte` | Pending | Hitung UTC. |
| M15 | Dialog tanpa focus trap / restore | `frontend/src/lib/Nav.svelte`, `ConfirmDialog.svelte` | Pending | Trap Tab + restore focus. |
| M16 | Refresh workspace race | `frontend/src/lib/workspace.svelte.js`, `Workspace.svelte` | Pending | Single in-flight/generation. |
| M17 | Terms tidak reset saat source berubah | `frontend/src/lib/CreateProject.svelte` | Pending | Reset `accepted`. |
| M18 | Status domain lama tampil setelah domain baru disimpan | `frontend/src/lib/ProjectDetail.svelte` | Pending | Reset state sebelum check. |
| M19 | API client tanpa timeout | `frontend/src/lib/api.js` | Pending | Default AbortController timeout. |
| M20 | Parsing ZIP freeze main thread | `frontend/src/lib/zipTree.js` | Pending | Debounce + worker/cap. |

## Low

| ID | Temuan | File utama | Status | Catatan |
|---|---|---|---|---|
| L1 | Href belum divalidasi scheme, `noopener` belum eksplisit | `frontend/src/lib/ProjectCard.svelte`, `ProjectDetail.svelte` | Pending | Allowlist http/https. |
| L2 | Tidak ada sign out di halaman publik saat sudah login | `frontend/src/lib/Nav.svelte`, `App.svelte` | Pending | Kirim handler logout ke semua Nav. |
| L3 | Clipboard error ditelan tanpa feedback | `frontend/src/lib/InstantDeploy.svelte`, `ProjectDetail.svelte` | Pending | Tampilkan gagal copy. |
| L4 | Tidak ada link project setelah deploy sukses | `frontend/src/lib/CreateProject.svelte` | Pending | Tampilkan link detail. |
| L5 | `DEV_LOGIN` GET state-changing tanpa origin guard | `backend/internal/handler/devlogin.go`, `router.go` | Pending | POST + CSRF / dev-only listener. |
| L6 | Tidak ada graceful HTTP shutdown | `backend/internal/server/process_manager.go`, `cmd/server/main.go` | Pending | `srv.Shutdown(ctx)`. |
| L7 | Logout hanya session saat ini | `backend/internal/handler/auth.go` | Pending | Opsi logout all sessions. |
| L8 | Callback OAuth tidak di-rate-limit | `backend/internal/server/router.go` | Pending | Tambah rate limit ringan. |
| L9 | Static fallback salah folder `.releases` | `backend/internal/server/sites.go` | Pending | Path `sitesDir/.releases`. |
| L10 | `ClearDeployments` meninggalkan folder release | `backend/internal/handler/project.go` | Pending | Hapus rilis tak terpakai. |

## Log Progres

- 2026-11-27 — Audit awal selesai; note dibuat. Belum ada perbaikan kode.
- 2026-11-27 — C1 done: env build/SSR diputus dari secret server; test + vet lulus.
- 2026-11-27 — C2 done: ZIP tidak bisa memicu SSR; test + vet lulus.
