"use client"

import { useEffect } from "react"
import { usePathname, useRouter } from "next/navigation"
import { useCurrentUser } from "@/lib/auth"
import styles from "./AuthGate.module.css"

/**
 * Membungkus halaman yang butuh sesi. Selama sesi dimuat, yang tampil hanya
 * layar tunggu (tanpa sidebar), sehingga navigasi tidak berkedip antara versi
 * tamu dan versi login. Kalau ternyata belum login, pengguna diarahkan ke
 * homepage (/) untuk sign in.
 */
export default function AuthGate({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const pathname = usePathname() || ""
  const { user, loaded } = useCurrentUser()

  const signedOut = loaded && !user

  useEffect(() => {
    if (!signedOut) return
    router.replace(`/`)
  }, [signedOut, router, pathname])

  if (!loaded || !user) {
    return (
      <div className={styles.screen}>
        <span className={styles.spinner} aria-hidden="true" />
        <span className={styles.label}>{signedOut ? "Redirecting to home…" : "Checking your session…"}</span>
      </div>
    )
  }

  return <>{children}</>
}
