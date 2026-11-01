"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import Sidebar from "./Sidebar"
import { User } from "@/lib/api"
import styles from "./Sidebar.module.css"

const PUBLIC_LINKS = [
  { href: "/docs", label: "Docs" },
  { href: "/help", label: "Help" },
  { href: "/support", label: "Support" },
]

export default function AppShell({
  user,
  loading = false,
  publicShell = false,
  children,
}: {
  user?: User | null
  loading?: boolean
  /** Halaman yang boleh dibaca tanpa login. Tamu melihat bar publik, user login tetap di dashboard. */
  publicShell?: boolean
  children: React.ReactNode
}) {
  const pathname = usePathname() || ""
  const authed = Boolean(user)

  // Selama sesi masih dimuat di halaman publik, tampilkan bar kosong (brand saja)
  // agar tidak ada pergantian sidebar <-> bar atas yang terlihat berkedip.
  if (publicShell && loading) {
    return (
      <div className={`${styles.shell} ${styles.publicShell}`}>
        <header className={styles.publicBar}>
          <Link href="/" className={styles.publicBrand}>
            only4.fun
          </Link>
          <nav className={styles.publicNav} />
        </header>
        <div className={styles.main}>{children}</div>
      </div>
    )
  }

  // Halaman publik untuk tamu: bar tipis standar di atas konten, tanpa sidebar dashboard.
  if (publicShell && !authed) {
    return (
      <div className={`${styles.shell} ${styles.publicShell}`}>
        <header className={styles.publicBar}>
          <Link href="/" className={styles.publicBrand}>only4.fun</Link>
          <nav className={styles.publicNav}>
            {PUBLIC_LINKS.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className={pathname.startsWith(link.href) ? styles.publicNavActive : ""}
              >
                {link.label}
              </Link>
            ))}
            <Link href="/login" className={styles.publicCta}>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.44 9.8 8.21 11.39.6.11.82-.26.82-.58 0-.29-.01-1.04-.02-2.04-3.34.73-4.04-1.61-4.04-1.61-.55-1.39-1.35-1.76-1.35-1.76-1.11-.76.08-.75.08-.75 1.23.09 1.88 1.28 1.88 1.28 1.09 1.87 2.87 1.34 3.57 1.03.11-.8.43-1.34.78-1.65-2.7-.31-5.54-1.36-5.54-6.02 0-1.32.47-2.4 1.24-3.24-.12-.31-.54-1.56.12-3.26 0 0 1.01-.33 3.3 1.25A11.5 11.5 0 0112 5.8c1 .004 2.01.14 2.91.41 2.28-1.58 3.29-1.25 3.29-1.25.66 1.7.24 2.95.12 3.26.78.85 1.23 1.92 1.23 3.24 0 4.68-2.85 5.71-5.56 5.98.44.38.82 1.14.82 2.3 0 1.66-.01 2.98-.01 3.38 0 .32.22.71.83.57A12.004 12.004 0 0024 12c0-6.63-5.37-12-12-12z" />
              </svg>
              Sign in
            </Link>
          </nav>
        </header>
        <div className={styles.main}>{children}</div>
      </div>
    )
  }

  // User login (atau halaman yang memang butuh sesi): tampilkan dashboard dengan sidebar.
  return (
    <div className={styles.shell}>
      <Sidebar user={user} loading={loading} />
      <div className={styles.main}>{children}</div>
    </div>
  )
}
