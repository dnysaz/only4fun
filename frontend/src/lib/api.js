// Klien API only4.fun. Semua request memakai cookie sesi (httpOnly) milik backend,
// jadi credentials harus selalu disertakan.

const BASE = import.meta.env.VITE_API_BASE ?? '';

export class ApiError extends Error {
  constructor(message, status = 0) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

async function request(path, { method = 'GET', body, signal } = {}) {
  const init = { method, credentials: 'include', signal, headers: {} };

  if (body instanceof FormData) {
    init.body = body;
  } else if (body !== undefined) {
    init.body = JSON.stringify(body);
    init.headers['Content-Type'] = 'application/json';
  }

  let res;
  try {
    res = await fetch(BASE + path, init);
  } catch (err) {
    if (err?.name === 'AbortError') throw err;
    throw new ApiError('Could not reach the API. Is the backend running?', 0);
  }

  const isJson = (res.headers.get('content-type') ?? '').includes('application/json');
  const data = isJson ? await res.json().catch(() => null) : null;

  if (!res.ok) {
    throw new ApiError(data?.error ?? `Request failed (${res.status})`, res.status);
  }

  return data;
}

export const api = {
  me: () => request('/api/auth/me'),
  logout: () => request('/api/auth/logout', { method: 'POST' }),

  listProjects: () => request('/api/projects'),
  getProject: (id) => request(`/api/projects/${encodeURIComponent(id)}`),
  createProject: ({ repoUrl, subdomain }) =>
    request('/api/projects', { method: 'POST', body: { repo_url: repoUrl, subdomain } }),
  uploadZip: (subdomain, file) => {
    const form = new FormData();
    form.append('subdomain', subdomain);
    form.append('file', file);
    return request('/api/projects/upload', { method: 'POST', body: form });
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

  checkSubdomain: (name) => request(`/api/check-subdomain?name=${encodeURIComponent(name)}`),

  // Instant deploy: publik, tanpa sesi, static HTML/CSS/JS saja.
  instantAddress: () => request('/api/instant/address', { method: 'POST' }),
  instantRepoTree: (repoUrl) =>
    request('/api/instant/repo-tree', { method: 'POST', body: { repo_url: repoUrl } }),
  instantRepo: (repoUrl, subdomain, addressToken) =>
    request('/api/instant/repo', {
      method: 'POST',
      body: { repo_url: repoUrl, subdomain, address_token: addressToken },
    }),
  instantZip: (file, subdomain, addressToken) => {
    const form = new FormData();
    form.append('file', file);
    if (subdomain) form.append('subdomain', subdomain);
    if (addressToken) form.append('address_token', addressToken);
    return request('/api/instant/zip', { method: 'POST', body: form });
  },

  loginUrl: (next = '/launchpad') => `${BASE}/auth/github/login?next=${encodeURIComponent(next)}`,
};
