// Data bersama untuk halaman /launchpad dan /deploy.

import { api } from './api.js';
import {
  invalidateSession,
  loadSession,
  session,
  setSessionInvalidationHandler,
} from './session.svelte.js';

export const workspace = $state({
  projects: [],
  loading: true,
  refreshing: false,
  error: '',
});

let started = false;
let reloadInFlight = null;
let reloadGeneration = 0;

setSessionInvalidationHandler(resetWorkspace);

export async function reload({ quiet = false } = {}) {
  // Poll 5 detik dan refresh manual bisa bersamaan. Pakai satu request dan
  // generation supaya response lama tidak menimpa state yang lebih baru.
  if (reloadInFlight) return reloadInFlight;

  const token = ++reloadGeneration;
  if (quiet) workspace.refreshing = true;
  else workspace.loading = true;
  workspace.error = '';

  const promise = (async () => {
    try {
      const data = await api.listProjects();
      if (token !== reloadGeneration) return;
      workspace.projects = data?.projects ?? [];
    } catch (err) {
      if (token !== reloadGeneration) return;
      if (err?.status === 401) {
        invalidateSession();
      } else {
        workspace.error = err?.message ?? 'Could not load your projects';
      }
    } finally {
      if (token === reloadGeneration) {
        workspace.loading = false;
        workspace.refreshing = false;
      }
    }
  })();

  reloadInFlight = promise;
  try {
    return await promise;
  } finally {
    if (reloadInFlight === promise) reloadInFlight = null;
  }
}

export async function ensureWorkspace() {
  if (started) return;
  started = true;

  try {
    await loadSession({ force: true });
    if (session.user) await reload();
    else workspace.loading = false;
  } catch {
    workspace.loading = false;
  }
}

export function resetWorkspace() {
  reloadGeneration += 1;
  reloadInFlight = null;
  started = false;
  workspace.projects = [];
  // Jangan biarkan splash loading menggantung setelah logout/expiry.
  workspace.loading = false;
  workspace.refreshing = false;
  workspace.error = '';
}
