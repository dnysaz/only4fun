export const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"

// Callback yang dipanggil setiap kali API menjawab 401. lib/auth.ts mendaftarkan
// handler di sini agar sesi yang kedaluwarsa langsung membersihkan cache user
// dan mengarahkan pengguna ke /login, tanpa menunggu reload.
type UnauthorizedHandler = () => void
let unauthorizedHandler: UnauthorizedHandler | null = null

export function setUnauthorizedHandler(handler: UnauthorizedHandler) {
  unauthorizedHandler = handler
}

// Session disimpan backend sebagai cookie httpOnly, jadi JavaScript tidak bisa
// (dan tidak perlu) membacanya. Semua request cukup mengirim kredensial.
export async function apiFetch(path: string, options?: RequestInit): Promise<Response> {
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    credentials: "include",
  })
  if (res.status === 401) {
    unauthorizedHandler?.()
  }
  return res
}

export interface User {
  id: string
  github_id: string
  username: string
  email: string
  avatar_url: string
}
