// Validasi URL repo GitHub di sisi klien. Aturannya mengikuti
// backend/internal/handler/project.go::validateRepoURL supaya tombol submit
// tidak aktif untuk URL yang pasti ditolak backend.

export function isValidGitHubRepoUrl(value) {
  return parseGitHubRepoUrl(value) !== null;
}

// Parse a github.com/owner/repo URL into the pieces used by the authenticated
// tree endpoint. Also keeps the client validation in sync with the backend.
export function parseGitHubRepoUrl(value) {
  const raw = String(value ?? '').trim();
  if (!raw) return null;

  try {
    const url = new URL(raw);
    if (url.protocol !== 'https:' && url.protocol !== 'http:') return null;
    if (url.username || url.password) return null;

    const host = url.hostname.toLowerCase();
    if (host !== 'github.com' && host !== 'www.github.com') return null;

    const parts = url.pathname.split('/').filter(Boolean);
    if (parts.length < 2) return null;

    const owner = parts[0];
    const repo = parts[1].replace(/\.git$/i, '');
    if (!/^[A-Za-z0-9_.-]+$/.test(owner) || !/^[A-Za-z0-9_.-]+$/.test(repo)) return null;
    if (owner === '.' || owner === '..' || repo === '.' || repo === '..') return null;

    return { owner, repo };
  } catch {
    return null;
  }
}
