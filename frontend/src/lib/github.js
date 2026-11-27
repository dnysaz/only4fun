// Validasi URL repo GitHub di sisi klien. Aturannya mengikuti
// backend/internal/handler/project.go::validateRepoURL supaya tombol submit
// tidak aktif untuk URL yang pasti ditolak backend.

export function isValidGitHubRepoUrl(value) {
  const raw = String(value ?? '').trim();
  if (!raw) return false;

  try {
    const url = new URL(raw);
    if (url.protocol !== 'https:' && url.protocol !== 'http:') return false;
    if (url.username || url.password) return false;

    const host = url.hostname.toLowerCase();
    return host === 'github.com' || host === 'www.github.com';
  } catch {
    return false;
  }
}
