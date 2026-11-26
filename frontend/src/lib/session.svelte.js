// Sesi user: dimuat sekali dari /api/auth/me, dipakai nav & dashboard.

import { api, ApiError } from './api.js';

export const session = $state({
  user: null,
  loading: true,
  error: '',
});

let started = false;

export async function loadSession({ force = false } = {}) {
  if (started && !force) return;
  started = true;
  session.loading = true;
  session.error = '';
  try {
    const data = await api.me();
    session.user = data?.user ?? null;
  } catch (err) {
    session.user = null;
    if (!(err instanceof ApiError) || err.status !== 401) {
      session.error = err?.message ?? 'Could not load your session';
    }
  } finally {
    session.loading = false;
  }
}

export async function signOut() {
  try {
    await api.logout();
  } catch {
    // Sesi mungkin sudah kedaluwarsa; tetap bersihkan state lokal.
  }
  session.user = null;
  started = false;
}
