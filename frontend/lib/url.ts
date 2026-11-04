// URL yang berasal dari API/database tidak boleh dipakai langsung sebagai href:
// nilai seperti `javascript:alert(1)` akan dieksekusi saat link diklik. Semua
// link eksternal harus lewat helper ini supaya hanya http(s) yang lolos.
const SAFE_PROTOCOLS = new Set(["http:", "https:"])

export function safeExternalUrl(raw: string | null | undefined): string | undefined {
  if (!raw) return undefined
  const value = raw.trim()
  if (!value) return undefined
  try {
    const parsed = new URL(value)
    if (!SAFE_PROTOCOLS.has(parsed.protocol)) return undefined
    return parsed.toString()
  } catch {
    return undefined
  }
}

// Website pada profil GitHub boleh ditulis tanpa skema ("example.com"), jadi
// tambahkan https:// lebih dulu lalu validasi seperti link eksternal lainnya.
export function safeWebsiteUrl(raw: string | null | undefined): string | undefined {
  if (!raw) return undefined
  const value = raw.trim().replace(/^\/+/, "")
  if (!value) return undefined
  if (/^https?:\/\//i.test(value)) return safeExternalUrl(value)
  return safeExternalUrl(`https://${value}`)
}
