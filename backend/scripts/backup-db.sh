#!/usr/bin/env bash
# Backup database SQLite only4.fun tanpa menghentikan server.
# Pemakaian:
#   DB_PATH=/var/lib/only4.fun/only4.fun.db ./scripts/backup-db.sh
# Jadwalkan lewat cron/systemd timer, misal tiap 6 jam:
#   0 */6 * * * /opt/only4.fun/backend/scripts/backup-db.sh
set -euo pipefail

DB_PATH="${DB_PATH:-$(cd "$(dirname "$0")/.." && pwd)/data/only4.fun.db}"
BACKUP_DIR="${BACKUP_DIR:-$(dirname "$DB_PATH")/backups}"
KEEP="${KEEP:-14}"
STAMP="$(date +%Y%m%d_%H%M%S)"
TARGET="$BACKUP_DIR/only4.fun.db.$STAMP"

if [ ! -f "$DB_PATH" ]; then
  echo "database tidak ditemukan: $DB_PATH" >&2
  exit 1
fi

mkdir -p "$BACKUP_DIR"

# .backup memakai API SQLite sehingga aman walau server sedang menulis.
if command -v sqlite3 >/dev/null 2>&1; then
  sqlite3 "$DB_PATH" ".backup '$TARGET'"
else
  cp "$DB_PATH" "$TARGET"
fi

# Simpan hanya N backup terbaru.
ls -t "$BACKUP_DIR"/only4.fun.db.* 2>/dev/null | tail -n "+$((KEEP + 1))" | xargs -r rm --

echo "backup dibuat: $TARGET"
