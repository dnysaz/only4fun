// Sesi user: dimuat sekali dari /api/auth/me, dipakai nav & dashboard.

import { api, ApiError } from './api.js';

export const session = $state({
  user: null,
  loading: true,
  error: '',
});

let started = false;
let inFlight = null;

export async function loadSession({ force = false } = {}) {
  if (started && !force) return inFlight;
  if (inFlight && !force) return inFlight;

  started = true;
  session.loading = true;
  session.error = '';

  inFlight = (async () => {
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
  })();

  try {
    return await inFlight;
  } finally {
    inFlight = null;
  }
}

// Cek ulang sesi tanpa mengguncang UI (visibility, route change, idle timer).
// Backend mengembalikan 401 bila cookie habis ATAU token GitHub sudah dicabut.
export async function revalidateSession() {
  if (inFlight) return inFlight;

  inFlight = (async () => {
    try {
      const data = await api.me();
      session.user = data?.user ?? null;
      session.error = '';
      return !!session.user;
    } catch (err) {
      if (err?.status === 401 || err?.status === 404) {
        session.user = null;
        session.error = '';
        started = false;
        return false;
      }
      // Gangguan jaringan bukan alasan untuk mengeluarkan user.
      return !!session.user;
    }
  })();

  try {
    return await inFlight;
  } finally {
    inFlight = null;
  }
}

export async function signOut() {
  try {
    await api.logout();
  } catch {
    // Sesi mungkin sudah kedaluwarsa; tetap bersihkan state lokal.
  }
  session.user = null;
  session.error = '';
  started = false;
}
