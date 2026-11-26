// Data bersama untuk halaman /launchpad dan /deploy.

import { api } from './api.js';
import { loadSession, session } from './session.svelte.js';

export const workspace = $state({
  projects: [],
  loading: true,
  refreshing: false,
  error: '',
});

let started = false;

export async function reload({ quiet = false } = {}) {
  if (quiet) workspace.refreshing = true;
  else workspace.loading = true;
  workspace.error = '';

  try {
    const data = await api.listProjects();
    workspace.projects = data?.projects ?? [];
  } catch (err) {
    if (err?.status === 401) {
      session.user = null;
    } else {
      workspace.error = err?.message ?? 'Could not load your projects';
    }
  } finally {
    workspace.loading = false;
    workspace.refreshing = false;
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
  started = false;
  workspace.projects = [];
  workspace.loading = true;
  workspace.refreshing = false;
  workspace.error = '';
}
