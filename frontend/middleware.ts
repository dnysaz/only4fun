import { NextResponse } from "next/server"
import type { NextRequest } from "next/server"

// Halaman yang butuh sesi. Sisanya publik: /, /login, /docs, /help, /support.
const PROTECTED_PREFIXES = ["/dashboard", "/projects", "/profile"]

// Middleware hanya memeriksa keberadaan cookie sesi (backend tetap yang
// memvalidasi di setiap request API). Tujuannya agar HTML halaman privat tidak
// pernah dikirim ke tamu — AuthGate di klien hanya jadi lapisan kedua.
export function middleware(request: NextRequest) {
  const { pathname, search } = request.nextUrl
  const isProtected = PROTECTED_PREFIXES.some(
    (prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`),
  )

  if (!isProtected) return NextResponse.next()
  if (request.cookies.get("session_id")?.value) return NextResponse.next()

  const loginUrl = request.nextUrl.clone()
  loginUrl.pathname = "/login"
  loginUrl.search = `?next=${encodeURIComponent(pathname + search)}`
  return NextResponse.redirect(loginUrl)
}

export const config = {
  matcher: ["/dashboard/:path*", "/projects/:path*", "/profile/:path*"],
}
