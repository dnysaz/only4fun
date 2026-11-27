// Klien API only4.fun. Semua request memakai cookie sesi (httpOnly) milik backend,
// jadi credentials harus selalu disertakan.

const BASE = import.meta.env.VITE_API_BASE ?? '';
const DEFAULT_TIMEOUT_MS = 20_000;

export class ApiError extends Error {
  constructor(message, status = 0) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

async function request(path, { method = 'GET', body, signal, timeoutMs = DEFAULT_TIMEOUT_MS } = {}) {
  const controller = new AbortController();
  let timedOut = false;

  // Caller-provided signals still abort the real request and surface AbortError.
  const relayAbort = () => controller.abort();
  if (signal?.aborted) relayAbort();
  else signal?.addEventListener('abort', relayAbort, { once: true });

  const timer =
    timeoutMs > 0
      ? setTimeout(() => {
          timedOut = true;
          controller.abort();
        }, timeoutMs)
      : null;

  const init = { method, credentials: 'include', signal: controller.signal, headers: {} };

  if (body instanceof FormData) {
    init.body = body;
  } else if (body !== undefined) {
    init.body = JSON.stringify(body);
    init.headers['Content-Type'] = 'application/json';
  }

  try {
    let res;
    try {
      res = await fetch(BASE + path, init);
    } catch (err) {
      if (err?.name === 'AbortError' && !timedOut) throw err;
      if (timedOut) throw new ApiError('The request took too long. Please try again.', 0);
      throw new ApiError('Could not reach the API. Is the backend running?', 0);
    }

    const isJson = (res.headers.get('content-type') ?? '').includes('application/json');
    const data = isJson ? await res.json().catch(() => null) : null;

    if (!res.ok) {
      throw new ApiError(data?.error ?? `Request failed (${res.status})`, res.status);
    }

    return data;
  } finally {
    if (timer) clearTimeout(timer);
    signal?.removeEventListener('abort', relayAbort);
  }
}

export const api = {
  me: () => request('/api/auth/me'),
  // all=true menghapus semua session user. Query dipakai supaya middleware
  // audit bisa mencatat "logout everywhere" tanpa membaca body request.
  logout: ({ all = false } = {}) =>
    request(`/api/auth/logout${all ? '?all=true' : ''}`, { method: 'POST' }),

  listProjects: () => request('/api/projects'),
  getProject: (id) => request(`/api/projects/${encodeURIComponent(id)}`),
  createProject: ({ repoUrl, subdomain, envVars }) =>
    request('/api/projects', {
      method: 'POST',
      body: {
        repo_url: repoUrl,
        subdomain,
        ...(envVars && Object.keys(envVars).length ? { env_vars: envVars } : {}),
      },
    }),
  uploadZip: (subdomain, file) => {
    const form = new FormData();
    form.append('subdomain', subdomain);
    form.append('file', file);
    // Upload sampai 100 MB butuh waktu lebih lama dari request biasa.
    return request('/api/projects/upload', { method: 'POST', body: form, timeoutMs: 120_000 });
  },
  deleteProject: (id) => request(`/api/projects/${encodeURIComponent(id)}`, { method: 'DELETE' }),
  redeploy: (id) => request(`/api/projects/${encodeURIComponent(id)}/deploy`, { method: 'POST' }),
  deployments: (id) => request(`/api/projects/${encodeURIComponent(id)}/deployments`),
  rollbackDeployment: (id, deploymentID) =>
    request(`/api/projects/${encodeURIComponent(id)}/deployments/${encodeURIComponent(deploymentID)}/rollback`, {
      method: 'POST',
    }),
  getEnv: (id) => request(`/api/projects/${encodeURIComponent(id)}/env`),
  updateEnv: (id, envVars) =>
    request(`/api/projects/${encodeURIComponent(id)}/env`, {
      method: 'POST',
      body: { env_vars: envVars },
    }),
  updateDomain: (id, domain) =>
    request(`/api/projects/${encodeURIComponent(id)}/domain`, {
      method: 'PUT',
      body: { domain },
    }),
  domainStatus: (id) => request(`/api/projects/${encodeURIComponent(id)}/domain/status`),

  checkSubdomain: (name) => request(`/api/check-subdomain?name=${encodeURIComponent(name)}`),
  listGitHubRepos: ({ page = 1, perPage = 100, signal } = {}) =>
    request(`/api/github/repos?page=${page}&per_page=${perPage}`, { signal }),

  // Instant deploy: publik, tanpa sesi, static HTML/CSS/JS saja.
  instantAddress: () => request('/api/instant/address', { method: 'POST' }),
  instantRepoTree: (repoUrl) =>
    request('/api/instant/repo-tree', { method: 'POST', body: { repo_url: repoUrl } }),
  instantRepo: (repoUrl, subdomain, addressToken) =>
    request('/api/instant/repo', {
      method: 'POST',
      body: { repo_url: repoUrl, subdomain, address_token: addressToken },
      timeoutMs: 120_000,
    }),
  instantZip: (file, subdomain, addressToken) => {
    const form = new FormData();
    form.append('file', file);
    if (subdomain) form.append('subdomain', subdomain);
    if (addressToken) form.append('address_token', addressToken);
    return request('/api/instant/zip', { method: 'POST', body: form, timeoutMs: 120_000 });
  },

  loginUrl: (next = '/launchpad') => `${BASE}/auth/github/login?next=${encodeURIComponent(next)}`,
};
