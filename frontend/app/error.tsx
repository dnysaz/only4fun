"use client"

import Link from "next/link"
import { useEffect } from "react"
import styles from "./not-found.module.css"

export default function GlobalRouteError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  useEffect(() => {
    // Simpan di konsol agar bisa dilaporkan tanpa menampilkan detail ke pengunjung.
    console.error("Route error:", error)
  }, [error])

  return (
    <main className={styles.page}>
      <div className={styles.grid} aria-hidden="true" />
      <div className={styles.content}>
        <Link href="/" className={styles.brand}>
          Only4fun
        </Link>
        <p className={styles.code}>Something broke</p>
        <h1 className={styles.title}>This page could not load.</h1>
        <p className={styles.lede}>
          The request failed on the way to the server. Retrying usually fixes it — if it keeps happening, the API may be
          down or your session may have expired.
        </p>
        <div className={styles.links}>
          <button type="button" onClick={reset} className={styles.link}>
            Try again
          </button>
          <Link href="/" className={styles.link}>
            Home
          </Link>
          <Link href="/" className={styles.link}>
            Sign in
          </Link>
        </div>
      </div>
    </main>
  )
}
