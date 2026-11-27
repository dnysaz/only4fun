# Production runbook — Only4fun

Dokumen ini untuk operator server. **Jangan menyimpan secret di sini** atau di
Git; semua nilai rahasia tetap berada di `.env`, `EnvironmentFile`, atau secret
manager. Ganti nama unit/systemd dan path contoh sesuai instalasi Anda.

## 1. Checklist sebelum traffic produksi

- [ ] `DEV_LOGIN=false`. Uji: `GET /auth/dev-login` harus `404`.
- [ ] `SESSION_SECRET` unik dan minimal 32 karakter, mis. `openssl rand -hex 32`.
      Server menolak start bila kosong/pendek saat `DEV_LOGIN=false`.
- [ ] `CORS_ORIGINS` hanya berisi origin frontend asli, mis.
      `https://only4.fun,https://www.only4.fun`. Jangan pakai `*`.
- [ ] `TRUSTED_PROXIES` hanya CIDR proxy yang benar-benar dipakai (loopback
      atau rentang IP Cloudflare). **Jangan** `0.0.0.0/0`: rate limit per-IP
      bergantung pada header yang hanya boleh datang dari proxy tepercaya.
- [ ] TLS di reverse proxy, HSTS aktif, dan `SERVER_ADDR` tidak terekspos
      publik. Firewall hanya membuka `80/443` dan SSH/admin.
- [ ] `FRONTEND_URL` dan `GITHUB_REDIRECT_URL` sesuai domain, dan callback
      terdaftar persis di GitHub OAuth App.
- [ ] Token Cloudflare hanya punya izin `Zone:DNS:Edit` dan
      `Zone:SSL and Certificates:Edit`. Cloudflare for SaaS + Fallback Origin
      aktif bila custom domain perlu SSL otomatis.
- [ ] `DATABASE_PATH`, `SITES_DIR`, dan `CACHE_DIR` bisa ditulis oleh user
      layanan, berada di disk dengan ruang cukup, dan tidak ikut tersaji sebagai
      file statis publik.
- [ ] `-workers` disesuaikan dengan CPU/RAM. Antrean build dibatasi 64 job.
- [ ] Backup SQLite terjadwal **dan** restore sudah diuji di lingkungan
      terpisah (bagian 5).
- [ ] `GET /ready` mengembalikan `{"status":"ready"}` sebelum load balancer
      menerima traffic.

## 2. Konfigurasi & rotasi secret

| Variable | Catatan produksi |
| --- | --- |
| `SESSION_SECRET` | Wajib, unik, ≥32 char. Dipakai untuk hash session, token GitHub, env var project, dan webhook secret. Lihat peringatan rotasi di bawah. |
| `DEV_LOGIN` | Harus `false`. Mode ini sengaja mengizinkan kunci fallback. |
| `CORS_ORIGINS` | Origin eksplisit; dipakai CORS dan origin guard. |
| `TRUSTED_PROXIES` | CIDR proxy/Cloudflare; jangan terlalu luas. |
| `FRONTEND_URL`, `GITHUB_REDIRECT_URL` | Harus konsisten dengan domain dan OAuth callback. |
| `GITHUB_CLIENT_SECRET` | Rahasia OAuth. Simpan hanya di environment layanan. |
| `CLOUDFLARE_API_KEY`, `CLOUDFLARE_ZONE_ID` | Kosongkan bila DNS otomatis tidak dipakai. |
| `CACHE_DIR`, `SITES_DIR`, `BUILD_TIMEOUT_SECONDS` | Lokasi/build workspace dan batas waktu satu build. |

### Rotasi `SESSION_SECRET`

Rotasi kunci ini **tidak transparan**:

1. Semua session langsung invalid; user harus login ulang.
2. Env var project dan webhook secret lama tidak bisa didekripsi dengan kunci
   baru. Setelah rotasi, minta user memasukkan ulang env var penting dan
   regenerate webhook secret GitHub bila dipakai.

Prosedur:

```bash
NEW_SECRET="$(openssl rand -hex 32)"
# Tulis NEW_SECRET ke secret manager / EnvironmentFile layanan — jangan ke Git.
sudo systemctl restart only4fun
curl -fsS http://127.0.0.1:8080/ready
```

Untuk rotasi rahasia lain:

- **GitHub OAuth client secret**: buat secret baru di GitHub App, perbarui env,
  restart layanan, lalu cabut secret lama.
- **Cloudflare API token**: buat token baru dengan izin sama, perbarui env,
  restart, verifikasi DNS/custom hostname, lalu cabut token lama.

## 3. Isolasi build & SSR

Build berjalan di host sebagai proses `git`/`npm` biasa. `internal/execenv`
hanya membatasi environment (tidak ada `os.Environ()` dan reserved key
ditolak); **ini bukan isolasi OS**. Repository yang di-build dianggap kode
yang dipercaya karena `npm install`/`npm run build` bisa menjalankan script
seperti user layanan.

Rekomendasi berlapis:

- Jalankan layanan dengan user non-root khusus. Berikan akses tulis hanya ke
  `DATABASE_PATH`, `SITES_DIR`, dan `CACHE_DIR`.
- Bila memungkinkan, jalankan build di container/gVisor/VM terpisah dengan:
  - tanpa Docker socket, tanpa mount host selain direktori build;
  - egress dibatasi hanya ke GitHub dan registry npm (atau mirror internal);
  - batas `CPUQuota`, `MemoryMax`, `TasksMax`, plus `BUILD_TIMEOUT_SECONDS`;
  - `NoNewPrivileges=true` dan `ProtectSystem=strict` pada unit systemd.
- `SERVER_ADDR` hanya boleh dijangkau reverse proxy. Blokir akses publik ke
  port SSR lokal (kisaran 4000–5000) lewat firewall; kode sudah memaksa
  `HOST`/`BIND`/`LISTEN_ADDRESS=127.0.0.1`, tetapi firewall tetap lapisan
  cadangan.
- Pastikan hasil build tidak mengandung dotfile sensitif: static handler
  menolak segmen `.env`, `.git`, `.releases`, kecuali `.well-known`.

## 4. Rate limit & anti-abuse

Rate limit bawaan berjalan **in-memory per proses**, bukan shared store. Jika
menjalankan lebih dari satu replica, kuota tidak terbagi dan harus ditambah
limiter bersama (mis. Redis) atau ditangani di edge.

Batas saat ini:

| Endpoint | Batas |
| --- | --- |
| `/api/instant/address` | 30/m per IP |
| `/api/instant/repo-tree` | 20/m per IP |
| `/api/instant/repo`, `/api/instant/zip` | 6/m per IP |
| `/auth/github/callback` | 20/m per IP |
| `POST /api/projects` | 10/m per session (fallback IP) |
| Redeploy / upload / env / domain | 6–20/m per session |
| `/webhooks/github` | 120/m |

Tambahan yang disarankan:

- Pasang Cloudflare WAF rate-limit/challenge untuk `/api/instant/*` dan
  `/auth/github/*`. Kecualikan atau izinkan IP GitHub untuk webhook agar HMAC
  tetap diproses.
- Pasang Turnstile di form `/instant` sebelum memanggil endpoint publik.
- Tambah kuota harian per IP/akun untuk deploy anonim.
- Pastikan `TRUSTED_PROXIES` benar supaya `c.ClientIP()` bukan IP proxy.
- Jangan memblokir `/health` dan `/ready` dari IP monitoring. `/metrics`
  sebaiknya tetap hanya di-loopback.

## 5. Backup & restore SQLite

Backup otomatis:

```bash
DB_PATH=/var/lib/only4.fun/only4.fun.db \
BACKUP_DIR=/var/lib/only4.fun/backups \
KEEP=14 \
  /opt/only4.fun/backend/scripts/backup-db.sh
```

Jadwalkan misalnya tiap 6 jam via cron:

```cron
0 */6 * * * DB_PATH=/var/lib/only4.fun/only4.fun.db BACKUP_DIR=/var/lib/only4.fun/backups /opt/only4.fun/backend/scripts/backup-db.sh
```

Pasang `sqlite3` di server. Script memakai `.backup` (aman saat server
menulis); fallback `cp` hanya boleh dipakai saat server dihentikan atau saat
development.

### Uji restore (wajib, minimal bulanan)

Lakukan di staging atau setelah menghentikan layanan. **Jangan menguji dengan
menimpa database produksi yang masih aktif.**

```bash
BACKUP=/var/lib/only4.fun/backups/only4.fun.db.20261127_000000
DB_PATH=/var/lib/only4.fun/only4.fun.db

# 1. Cek integritas backup sebelum menyentuh live DB.
sqlite3 "$BACKUP" 'PRAGMA integrity_check;'   # harus: ok

# 2. Hentikan layanan dan simpan keepaside DB + WAL/SHM.
sudo systemctl stop only4fun
STAMP="$(date +%Y%m%d_%H%M%S)"
mkdir -p "$DB_PATH.keepaside"
mv "$DB_PATH" "$DB_PATH.keepaside/$STAMP.db"
[ -f "$DB_PATH-wal" ] && mv "$DB_PATH-wal" "$DB_PATH.keepaside/$STAMP.db-wal"
[ -f "$DB_PATH-shm" ] && mv "$DB_PATH-shm" "$DB_PATH.keepaside/$STAMP.db-shm"

# 3. Restore dari backup (sqlite3 menimpa file tujuan).
sqlite3 "$BACKUP" ".backup '$DB_PATH'"

# 4. Start lagi dan verifikasi.
sudo systemctl start only4fun
curl -fsS http://127.0.0.1:8080/ready
sqlite3 "$DB_PATH" 'PRAGMA integrity_check;'
```

Setelah restore, cek login, daftar project, env var, custom domain, dan riwayat
deployment. Backup hanya mencakup database; **`SITES_DIR` perlu di-backup atau
dibangun ulang** dari repository sumber bila terjadi kehilangan disk.

## 6. Observability

- `/health` = liveness (proses hidup). `/ready` = readiness (DB ping, worker,
  folder DB/sites/cache bisa ditulis). Pakai `/ready` untuk load balancer.
- `/metrics` hanya bisa diakses dari loopback. Contoh scrape lokal:

  ```bash
  curl -fsS http://127.0.0.1:8080/metrics
  ```

  Metrik meliputi `requests_total`, `errors_total`, `build_queued`,
  `build_active`, `build_total`, `build_succeeded`, `build_failed`,
  `build_last_duration_ms`, `build_last_failure`, dan `disk_free_*_bytes`.
- Setiap response punya header `X-Request-ID`; log JSON juga memuat
  `request_id`. Sertakan ID ini saat lapor insiden.
- Audit aksi sensitif tersimpan di tabel `audit_logs`:

  ```bash
  sqlite3 "$DB_PATH" \
    "SELECT created_at, action, target, ip, status FROM audit_logs ORDER BY created_at DESC LIMIT 20;"
  ```

## 7. Runbook insiden singkat

| Gejala | Langkah |
| --- | --- |
| `/ready` `503` | Cek alasan di body, izin folder, ruang disk, dan apakah worker start (`-workers`). Lihat `journalctl -u only4fun -n 200`. |
| Build stuck / gagal massal | Cek `build_active` dan `build_last_failure` di `/metrics`, log deployment di UI, serta akses ke GitHub/npm. Antrean penuh = 64 job. Redeploy saat `building`/`pending` memang ditolak `409`. |
| Situs SSR mati | Cek proses anak dan `ss -ltnp` (port harus loopback), log proxy, lalu restart layanan atau redeploy project. |
| Deploy anonim disalahgunakan | Cek `TRUSTED_PROXIES`, aktifkan WAF/Turnstile, blokir IP di edge, dan tinjau `audit_logs`/log JSON. |
| Custom domain SSL tidak aktif | Cek CNAME ke target `VITE_CUSTOM_DOMAIN_CNAME` (`customers.only4.fun`), Cloudflare for SaaS, izin token, lalu tombol **Check again**. Status `Waiting for DNS`/`Issuing certificate` wajar sebelum DNS propagate. |
| Disk penuh | Prune release otomatis menyimpan 3 rilis terakhir; bersihkan `CACHE_DIR` lama dan cek retensi backup. |
| Setelah rotasi `SESSION_SECRET` | User login ulang, input ulang env var, regenerate webhook secret bila dipakai. |

## 8. Kontak & eskalasi

Gunakan halaman `/support` untuk kanal bantuan. Sertakan: request ID,
subdomain/project ID, waktu kejadian (UTC), dan langkah yang sudah dicoba.
Jangan pernah mengirim token, cookie session, atau isi env var ke support.
