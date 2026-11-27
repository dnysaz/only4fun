// Sesi user: dimuat sekali dari /api/auth/me, dipakai nav & dashboard.

import { api } from './api.js';

export const session = $state({
  user: null,
  loading: true,
  error: '',
});

let started = false;
let inFlight = null;
let generation = 0;
let invalidateHandler = () => {};

// Workspace mendaftarkan resetWorkspace di sini supaya expiry sesi selalu
// membersihkan data project tanpa impor sirkular dari session ke workspace.
export function setSessionInvalidationHandler(handler) {
  invalidateHandler = typeof handler === 'function' ? handler : () => {};
}

// Satu jalur untuk semua kondisi "sesi tidak berlaku": logout, 401, dan 404.
// Counter membuat response lama (mis. revalidate yang sedang berjalan saat
// logout) tidak bisa mengembalikan session.user setelah state dibersihkan.
export function invalidateSession() {
  generation += 1;
  started = false;
  inFlight = null;
  session.loading = false;
  session.user = null;
  session.error = '';
  invalidateHandler();
}

function isStale(token) {
  return token !== generation;
}

export async function loadSession({ force = false } = {}) {
  if (started && !force) return inFlight;
  if (inFlight && !force) return inFlight;

  const token = generation;
  started = true;
  session.loading = true;
  session.error = '';

  const promise = (async () => {
    try {
      const data = await api.me();
      if (isStale(token)) return;
      session.user = data?.user ?? null;
    } catch (err) {
      if (isStale(token)) return;
      session.user = null;
      if (err?.status === 401 || err?.status === 404) {
        invalidateSession();
        return;
      }
      session.error = err?.message ?? 'Could not load your session';
    } finally {
      if (!isStale(token)) session.loading = false;
    }
  })();

  inFlight = promise;
  try {
    return await promise;
  } finally {
    if (inFlight === promise) inFlight = null;
  }
}

// Cek ulang sesi tanpa mengguncang UI (visibility, route change, idle timer).
// Backend mengembalikan 401 bila cookie habis ATAU token GitHub sudah dicabut.
export async function revalidateSession() {
  if (inFlight) return inFlight;

  const token = generation;
  const promise = (async () => {
    try {
      const data = await api.me();
      if (isStale(token)) return false;
      session.user = data?.user ?? null;
      session.error = '';
      return !!session.user;
    } catch (err) {
      if (isStale(token)) return false;
      if (err?.status === 401 || err?.status === 404) {
        invalidateSession();
        return false;
      }
      // Gangguan jaringan bukan alasan untuk mengeluarkan user.
      return !!session.user;
    }
  })();

  inFlight = promise;
  try {
    return await promise;
  } finally {
    if (inFlight === promise) inFlight = null;
  }
}

export async function signOut() {
  try {
    await api.logout();
  } catch {
    // Sesi mungkin sudah kedaluwarsa; tetap bersihkan state lokal.
  }
  invalidateSession();
}
