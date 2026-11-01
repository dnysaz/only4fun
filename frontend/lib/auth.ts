"use client"

import { useEffect, useState } from "react"
import { apiFetch, setUnauthorizedHandler, User } from "./api"

// Cache user di memori selama tab ini hidup. Tujuannya supaya perpindahan
// antarhalaman (mis. Docs -> Help) tidak menampilkan bar publik sebentar
// sebelum sidebar dashboard muncul lagi.
let cachedUser: User | null | undefined
let inflight: Promise<User | null> | null = null

// Halaman yang boleh dibuka tanpa sesi; 401 di halaman ini tidak boleh
// memicu redirect agar tamu tetap bisa membaca docs/help/support.
const PUBLIC_PATHS = ["/", "/login", "/docs", "/help", "/support"]

function isPublicPath(pathname: string) {
  return PUBLIC_PATHS.some((path) => pathname === path || pathname.startsWith(`${path}/`))
}

export function getCachedUser(): User | null | undefined {
  return cachedUser
}

/** Dipakai saat logout: pastikan halaman berikutnya tidak memakai user lama. */
export function clearCurrentUser() {
  cachedUser = null
}

// Sesi berakhir di server (24 jam atau kunci enkripsi berubah) sementara tab
// masih terbuka: request API berikutnya menjawab 401, cache dibuang, dan
// pengguna dikirim ke halaman login bersama tujuan asalnya.
setUnauthorizedHandler(() => {
  cachedUser = null
  if (typeof window === "undefined") return
  const { pathname, search } = window.location
  if (isPublicPath(pathname)) return
  window.location.assign(`/login?next=${encodeURIComponent(pathname + search)}`)
})

export function loadCurrentUser(): Promise<User | null> {
  if (inflight) return inflight
  const request = apiFetch("/api/auth/me")
    .then((res) => (res.ok ? res.json() : null))
    .then((data) => {
      const user: User | null = data?.user ?? null
      cachedUser = user
      return user
    })
    .catch(() => {
      cachedUser = null
      return null
    })
    .finally(() => {
      inflight = null
    })
  inflight = request
  return request
}

/**
 * Membaca user aktif. Kalau sebelumnya sudah pernah dimuat, nilai cache
 * langsung dipakai sehingga shell dashboard tampil tanpa kedip.
 */
export function useCurrentUser() {
  const [user, setUser] = useState<User | null>(() => cachedUser ?? null)
  const [loaded, setLoaded] = useState(() => cachedUser !== undefined)

  useEffect(() => {
    let active = true
    loadCurrentUser()
      .then((next) => {
        if (active) setUser(next)
      })
      .finally(() => {
        if (active) setLoaded(true)
      })
    return () => {
      active = false
    }
  }, [])

  return { user, loaded }
}
