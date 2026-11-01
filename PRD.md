# PRD — only4.fun

Static Site Hosting / Serverless (sederhana, mirip Vercel) dengan login GitHub, deploy dari Git repo, env vars, dan subdomain otomatis.

---

## 1. Tujuan

Membuat platform sederhana agar user bisa onlinekan situs statis/serverless dari Git repo, login via GitHub, dan langsung dapat subdomain `*.only4.fun`.

---

## 2. Pengguna (Personas)

| Tipe | Kebutuhan |
|------|-----------|
| User | Login GitHub, paste URL repo Git, isi env vars, dapatkan subdomain, lihat status deploy |
| Admin (opsional) | Lihat daftar project, kontrol quota, logout user |

---

## 3. Fitur Utama (MVP)

### 3.1 Autentikasi
- Login hanya melalui GitHub OAuth.
- Session disimpan di backend (cookie/JWT).
- Frontend menampilkan "Login with GitHub".

### 3.2 Deploy dari Git Repo
- User paste URL repo Git (publc/private).
- Backend clone repo.
- Deteksi framework/build command (opsional default: `npm run build` / `npm run build` dengan output `dist` atau `out`).
- Build di Docker container (isolasi).
- Hasil build disimpan di shared storage.

### 3.3 Environment Variables
- User input key:value di UI sebelum/kurang lebih saat deploy.
- ENV di-inject ke dalam proses build.
- Tampilkan daftar env yang sudah di-set per project.

### 3.4 Subdomain Otomatis
- Setiap project dapat subdomain: `namaproject.only4.fun`.
- DNS ditambahkan otomatis ke Cloudflare (via API).
- Nginx reverse proxy diterapkan untuk routing subdomain.

### 3.5 Dashboard
- Daftar project yang di-deploy.
- Status deploy (pending/building/deployed/failed).
- Tombol redeploy, delete project, edit env.

---

## 4. Arsitektur Teknis

```
Frontend        Backend API         Infrastructure
(Next.js)       (Go + Gin)          (Docker + Nginx + Cloudflare)
   │                  │                      │
   └─► GitHub OAuth ─► Repo Clone ─► Build ─► Serve via Nginx
   └─► Env Manager ─► Build Queue ─► Static Files ─► *.only4.fun
```

| Layer | Tech Stack | Tugas |
|-------|-----------|-------|
| Frontend | Next.js (App Router) | UI login, dashboard, form deploy, env input |
| Backend | Go + Gin | API, GitHub OAuth, build orchestrator |
| Builder | Docker | Clone repo, build static assets |
| Storage | Local disk / shared volume | Hasil build static files |
| Web Server | Nginx | Reverse proxy per subdomain |
| DNS | Cloudflare API | Otomatis buat A/CNAME record `*.only4.fun` |

---

## 5. Alur Deploy (User Journey)

1. User buka `only4.fun`, klik "Login with GitHub".
2. Redirect ke GitHub OAuth, kembali ke `only4.fun/dashboard`.
3. User klik "New Project", isi:
   - URL repo Git
   - Nama project (jadi subdomain)
   - Environment variables (key:value, optional)
4. Backend clone repo → build dengan Docker → simpan static files.
5. Backend tambahkan DNS record di Cloudflare untuk `namaproject.only4.fun`.
6. Nginx reload config, site online.
7. User buka `namaproject.only4.fun`, lihat situsnya.

---

## 6. API Endpoints (Backend Go Gin)

| Endpoint | Method | Tugas |
|----------|--------|-------|
| `/auth/github/login` | GET | Redirect ke GitHub OAuth |
| `/auth/github/callback` | GET | Handle callback, buat session |
| `/api/projects` | GET | Daftar project user |
| `/api/projects` | POST | Buat project baru (trigger deploy) |
| `/api/projects/:id` | GET | Detail project + status |
| `/api/projects/:id/deploy` | POST | Redeploy |
| `/api/projects/:id/env` | GET/POST/PATCH | Kelola env vars |
| `/api/projects/:id` | DELETE | Hapus project |
| `/api/auth/me` | GET | Info user login |

---

## 7. Non-Fungsional

- **Build isolation**: Setiap build di container terpisah (Docker).
- **Subdomain wildcard**: `*.only4.fun` di Cloudflare.
- **Rate limit**: Hindari abuse clone repo berulang.
- **Health check**: Endpoint `/health` untuk monitoring.

---

## 8. Out of Scope (v1)

- Custom domain (user beli sendiri).
- CI/CD otomatis pada push ke GitHub (webhook).
- Database production-grade (bisa mulai SQLite/boltdb dulu).
- Multi-region / auto-scaling.
- Server-side rendering / dynamic API routes (hanya static).

---

## 9. Timeline (Rencana Awal)

| Fase | Tugas | Estimasi |
|------|-------|----------|
| 1 | Setup repo, Go Gin backend dasar, GitHub OAuth | 1-2 hari |
| 2 | Frontend Next.js login + dashboard dasar | 1-2 hari |
| 3 | Build pipeline (Docker clone + build static) | 2-3 hari |
| 4 | Subdomain + Cloudflare DNS + Nginx routing | 1-2 hari |
| 5 | Env vars, redeploy, delete, polishing | 1-2 hari |
| **Total** | **MVP siap coba** | **6-11 hari** |

---

## 10. Risiko & Mitigasi

| Risiko | Mitigasi |
|--------|----------|
| GitHub OAuth rate limit | Cache token, jangan panggil API berlebihan |
| Build gagal / timeout | Batasi waktu build, tampilkan log error ke user |
| Subdomain konflik | Validasi nama project unik di backend |
| Keamanan clone repo private | Gunakan GitHub token atau deploy key |
| Nginx reload sering | Debounce reload, reload sekali per batch deploy |

---

*Dibuat sebagai referensi bersama saat membangun only4.fun. Dokumen ini bisa diperbarui seiring perkembangan.*