// Hanya izinkan URL absolut http/https untuk atribut href yang bisa diklik.

export function safeHttpUrl(value) {
  if (typeof value !== 'string' || !value.trim()) return '';
  const candidate = value.trim();
  try {
    const parsed = new URL(candidate);
    // Return the original spelling so the visible address is unchanged.
    return parsed.protocol === 'http:' || parsed.protocol === 'https:' ? candidate : '';
  } catch {
    return '';
  }
}
