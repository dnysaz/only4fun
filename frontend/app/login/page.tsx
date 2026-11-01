"use client"

import { useEffect, useState } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import { useCurrentUser } from "@/lib/auth"
import styles from "./login.module.css"

// Hanya izinkan tujuan internal agar ?next= tidak bisa dipakai untuk
// mengalihkan pengguna ke situs lain setelah login.
function safeNext(value: string | null) {
  if (!value) return "/dashboard"
  if (!value.startsWith("/") || value.startsWith("//")) return "/dashboard"
  return value
}

export default function LoginPage() {
  const router = useRouter()
  const { user, loaded } = useCurrentUser()
  const [next, setNext] = useState("/dashboard")

  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    setNext(safeNext(params.get("next")))
  }, [])

  // Sudah punya sesi: tidak perlu melihat halaman ini lagi.
  useEffect(() => {
    if (loaded && user) router.replace(next)
  }, [loaded, user, next, router])

  return (
    <main className={styles.page}>
      <div className={styles.grid} aria-hidden="true" />
      <div className={styles.glow} aria-hidden="true" />

      <div className={styles.content}>
        <Link href="/" className={styles.brand}>
          only4.fun
        </Link>

        <h1 className={styles.title}>Sign in to continue.</h1>
        <p className={styles.lede}>
          The dashboard, your projects and settings need a GitHub account. Docs, Help and Support stay open to
          everyone.
        </p>

        <a href={`/auth/github/login?next=${encodeURIComponent(next)}`} className={styles.cta}>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
            <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.44 9.8 8.21 11.39.6.11.82-.26.82-.58 0-.29-.01-1.04-.02-2.04-3.34.73-4.04-1.61-4.04-1.61-.55-1.39-1.35-1.76-1.35-1.76-1.11-.76.08-.75.08-.75 1.23.09 1.88 1.28 1.88 1.28 1.09 1.87 2.87 1.34 3.57 1.03.11-.8.43-1.34.78-1.65-2.7-.31-5.54-1.36-5.54-6.02 0-1.32.47-2.4 1.24-3.24-.12-.31-.54-1.56.12-3.26 0 0 1.01-.33 3.3 1.25A11.5 11.5 0 0112 5.8c1 .004 2.01.14 2.91.41 2.28-1.58 3.29-1.25 3.29-1.25.66 1.7.24 2.95.12 3.26.78.85 1.23 1.92 1.23 3.24 0 4.68-2.85 5.71-5.56 5.98.44.38.82 1.14.82 2.3 0 1.66-.01 2.98-.01 3.38 0 .32.22.71.83.57A12.004 12.004 0 0024 12c0-6.63-5.37-12-12-12z" />
          </svg>
          Continue with GitHub
        </a>

        <Link href="/" className={styles.back}>
          ← Back to home
        </Link>
      </div>
    </main>
  )
}
