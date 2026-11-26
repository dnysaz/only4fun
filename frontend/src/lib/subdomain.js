// Aturan nama subdomain, disamakan dengan backend (internal/handler/project.go).
import { isBlockedWord } from './check_word.ts';

export const RESERVED = [
  'www',
  'api',
  'app',
  'admin',
  'auth',
  'docs',
  'help',
  'profile',
  'dashboard',
  'launchpad',
  'mail',
];

export const MIN_LENGTH = 3;
export const MAX_LENGTH = 30;

export function sanitize(value) {
  return String(value ?? '')
    .toLowerCase()
    .replace(/[^a-z0-9-]/g, '')
    .slice(0, MAX_LENGTH);
}

/** @returns {{ ok: boolean, reason: '' | 'empty' | 'short' | 'hyphen' | 'blocked' | 'reserved' }} */
export function validateName(value) {
  if (!value) return { ok: false, reason: 'empty' };
  if (value.length < MIN_LENGTH) return { ok: false, reason: 'short' };
  if (/^-|-$/.test(value)) return { ok: false, reason: 'hyphen' };
  // Exact match: "admin" ditolak, "madmin" tetap boleh.
  if (isBlockedWord(value)) return { ok: false, reason: 'blocked' };
  if (RESERVED.includes(value)) return { ok: false, reason: 'reserved' };
  return { ok: true, reason: '' };
}
