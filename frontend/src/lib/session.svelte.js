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

// Cek ulang sesi tanpa mengguncang UI (dipakai saat tab dibuka kembali).
// Backend mengembalikan 401 bila cookie habis ATAU token GitHub sudah dicabut.
export async function revalidateSession() {
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
