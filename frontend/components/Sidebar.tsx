"use client"

import { useEffect, useState } from "react"
import Link from "next/link"
import { usePathname, useRouter } from "next/navigation"
import { apiFetch, User } from "@/lib/api"
import { clearCurrentUser } from "@/lib/auth"
import styles from "./Sidebar.module.css"

interface NavItem {
  label: string
  href: string
  icon: React.ReactNode
  match?: (pathname: string) => boolean
}

const iconProps = {
  width: 16,
  height: 16,
  viewBox: "0 0 24 24",
  fill: "none",
  stroke: "currentColor",
  strokeWidth: 1.8,
  strokeLinecap: "round" as const,
  strokeLinejoin: "round" as const,
}

const items: NavItem[] = [
  {
    label: "Dashboard",
    href: "/dashboard",
    icon: (
      <svg {...iconProps}>
        <rect x="3" y="3" width="7" height="7" rx="1.5" />
        <rect x="14" y="3" width="7" height="7" rx="1.5" />
        <rect x="3" y="14" width="7" height="7" rx="1.5" />
        <rect x="14" y="14" width="7" height="7" rx="1.5" />
      </svg>
    ),
    match: (p) => p === "/dashboard" || (p.startsWith("/projects") && p !== "/projects/new"),
  },
  {
    label: "New project",
    href: "/projects/new",
    icon: (
      <svg {...iconProps}>
        <path d="M12 5v14M5 12h14" />
      </svg>
    ),
    match: (p) => p === "/projects/new",
  },
  {
    label: "Docs",
    href: "/docs",
    icon: (
      <svg {...iconProps}>
        <path d="M4 5.5A2.5 2.5 0 0 1 6.5 3H20v15H6.5A2.5 2.5 0 0 0 4 20.5z" />
        <path d="M8 7h8M8 11h6" />
      </svg>
    ),
    match: (p) => p.startsWith("/docs"),
  },
  {
    label: "Help",
    href: "/help",
    icon: (
      <svg {...iconProps}>
        <circle cx="12" cy="12" r="9" />
        <path d="M9.5 9a2.5 2.5 0 1 1 3.4 2.3c-.6.3-.9.8-.9 1.4v.4" />
        <path d="M12 17h.01" />
      </svg>
    ),
    match: (p) => p.startsWith("/help"),
  },
  {
    label: "Support",
    href: "/support",
    icon: (
      <svg {...iconProps}>
        <path d="M12 21s-7-4.35-7-10a4 4 0 0 1 7-2.65A4 4 0 0 1 19 11c0 5.65-7 10-7 10z" />
      </svg>
    ),
    match: (p) => p.startsWith("/support"),
  },
  {
    label: "Profile",
    href: "/profile",
    icon: (
      <svg {...iconProps}>
        <circle cx="12" cy="12" r="3" />
        <path d="M19.4 15a1.6 1.6 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.6 1.6 0 0 0-1.8-.3 1.6 1.6 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1a1.6 1.6 0 0 0-1-1.5 1.6 1.6 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.6 1.6 0 0 0 .3-1.8 1.6 1.6 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1a1.6 1.6 0 0 0 1.5-1 1.6 1.6 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.6 1.6 0 0 0 1.8.3H9a1.6 1.6 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.6 1.6 0 0 0 1 1.5 1.6 1.6 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.6 1.6 0 0 0-.3 1.8V9a1.6 1.6 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.6 1.6 0 0 0-1.5 1z" />
      </svg>
    ),
    match: (p) => p.startsWith("/profile"),
  },
]

// Halaman publik (docs, help, support) bisa dibaca tanpa login, jadi sidebar
// menyesuaikan diri: tamu hanya melihat menu yang tidak butuh sesi.
const PUBLIC_HREFS = new Set(["/docs", "/help", "/support"])

export default function Sidebar({ user, loading = false }: { user?: User | null; loading?: boolean }) {
  const pathname = usePathname() || ""
  const router = useRouter()
  const [loggingOut, setLoggingOut] = useState(false)
  // Selalu mulai dari expanded agar render pertama klien sama dengan HTML
  // server; preferensi dari localStorage baru diterapkan setelah mount.
  const [collapsed, setCollapsed] = useState(false)
  // Di layar kecil sidebar jadi drawer, jadi ia tidak boleh mewarisi status
  // collapsed dari desktop: drawer selalu tampil dengan label.
  const [isMobile, setIsMobile] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)
  const authed = Boolean(user)
  const navItems = authed || loading ? items : items.filter((item) => PUBLIC_HREFS.has(item.href))
  const showLabels = !collapsed || isMobile

  useEffect(() => {
    if (localStorage.getItem("sidebar-collapsed") === "true") setCollapsed(true)
  }, [])

  useEffect(() => {
    const mq = window.matchMedia("(max-width: 900px)")
    const sync = () => {
      setIsMobile(mq.matches)
      if (!mq.matches) setMobileOpen(false)
    }
    sync()
    mq.addEventListener("change", sync)
    return () => mq.removeEventListener("change", sync)
  }, [])

  // Tutup drawer setiap kali pindah halaman.
  useEffect(() => {
    setMobileOpen(false)
  }, [pathname])

  // Kunci scroll halaman selagi drawer terbuka agar latar tidak ikut bergulir.
  useEffect(() => {
    if (!mobileOpen) return
    const previous = document.body.style.overflow
    document.body.style.overflow = "hidden"
    return () => {
      document.body.style.overflow = previous
    }
  }, [mobileOpen])

  const toggleCollapsed = () => {
    setCollapsed((prev) => {
      const next = !prev
      if (typeof window !== "undefined") {
        localStorage.setItem("sidebar-collapsed", String(next))
      }
      return next
    })
  }

  const handleLogout = async () => {
    setLoggingOut(true)
    try {
      await apiFetch("/api/auth/logout", { method: "POST" })
    } catch {
      // tetap arahkan ke beranda walau backend tidak merespons
    }
    clearCurrentUser()
    router.push("/login")
    router.refresh()
  }

  return (
    <>
      {/* Bar tipis khusus mobile: tombol buka menu + brand. */}
      <div className={styles.mobileBar}>
        <button
          type="button"
          className={styles.mobileToggle}
          onClick={() => setMobileOpen(true)}
          aria-label="Open menu"
          aria-expanded={mobileOpen}
          aria-controls="app-sidebar"
        >
          <svg {...iconProps}>
            <path d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        <Link href="/" className={styles.brand}>
          Only4fun
        </Link>
        {user && (
          <Link href="/profile" className={styles.mobileAvatar} aria-label={user.username}>
            {user.avatar_url ? (
              <img src={user.avatar_url} alt="" width={24} height={24} loading="lazy" decoding="async" className={styles.avatar} />
            ) : (
              <span className={styles.avatar} style={{ background: "#eaeaea", display: "inline-block" }} />
            )}
          </Link>
        )}
      </div>

      {mobileOpen && (
        <button type="button" className={styles.backdrop} onClick={() => setMobileOpen(false)} aria-label="Close menu" />
      )}

      <aside
        id="app-sidebar"
        className={`${styles.sidebar} ${collapsed && !isMobile ? styles.sidebarCollapsed : ""} ${mobileOpen ? styles.sidebarMobileOpen : ""}`}
      >
        <div className={styles.sidebarHeader}>
          {showLabels && (
            <Link href="/" className={styles.brand}>
              Only4fun
            </Link>
          )}
          <button
            onClick={() => (isMobile ? setMobileOpen(false) : toggleCollapsed())}
            className={styles.toggleBtn}
            aria-label={isMobile ? "Close menu" : collapsed ? "Expand sidebar" : "Collapse sidebar"}
          >
            {showLabels ? (
              <svg width={15} height={15} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.9} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                <path d="M19 12H5" />
                <path d="M11 18l-6-6 6-6" />
              </svg>
            ) : (
              <svg width={15} height={15} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.9} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                <path d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            )}
          </button>
        </div>

        <nav className={styles.nav}>
          {showLabels && <div className={styles.section}>Menu</div>}
          {navItems.map((item) => {
            const active = item.match ? item.match(pathname) : false
            return (
              <Link
                key={item.label}
                href={item.href}
                onClick={() => setMobileOpen(false)}
                className={`${styles.item} ${active ? styles.active : ""}`}
                title={showLabels ? undefined : item.label}
              >
                {item.icon}
                {showLabels && <span>{item.label}</span>}
              </Link>
            )
          })}

          {/* Menu tambahan, dipisahkan garis dari menu utama. */}
          <div className={styles.divider} role="separator" />
          <a
            href="https://moogo.dev"
            target="_blank"
            rel="noreferrer"
            className={`${styles.item} ${styles.externalItem}`}
            title="Moogo — SQLite over HTTP"
          >
            <svg {...iconProps}>
              <ellipse cx="12" cy="5" rx="7" ry="3" />
              <path d="M5 5v7c0 1.66 3.13 3 7 3s7-1.34 7-3V5" />
              <path d="M5 12v7c0 1.66 3.13 3 7 3s7-1.34 7-3v-7" />
            </svg>
            {showLabels && <span>Moogo</span>}
            {showLabels && (
              <svg width={12} height={12} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" className={styles.externalArrow} aria-hidden="true">
                <path d="M7 17 17 7" />
                <path d="M9 7h8v8" />
              </svg>
            )}
          </a>
        </nav>

        <div className={styles.spacer} />

        {!loading && !authed && (
          <Link href="/login" className={`${styles.item} ${styles.signIn}`} title={showLabels ? undefined : "Sign in with GitHub"}>
            <svg width={16} height={16} viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
              <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.44 9.8 8.21 11.39.6.11.82-.26.82-.58 0-.29-.01-1.04-.02-2.04-3.34.73-4.04-1.61-4.04-1.61-.55-1.39-1.35-1.76-1.35-1.76-1.11-.76.08-.75.08-.75 1.23.09 1.88 1.28 1.88 1.28 1.09 1.87 2.87 1.34 3.57 1.03.11-.8.43-1.34.78-1.65-2.7-.31-5.54-1.36-5.54-6.02 0-1.32.47-2.4 1.24-3.24-.12-.31-.54-1.56.12-3.26 0 0 1.01-.33 3.3 1.25A11.5 11.5 0 0112 5.8c1 .004 2.01.14 2.91.41 2.28-1.58 3.29-1.25 3.29-1.25.66 1.7.24 2.95.12 3.26.78.85 1.23 1.92 1.23 3.24 0 4.68-2.85 5.71-5.56 5.98.44.38.82 1.14.82 2.3 0 1.66-.01 2.98-.01 3.38 0 .32.22.71.83.57A12.004 12.004 0 0024 12c0-6.63-5.37-12-12-12z" />
            </svg>
            {showLabels && <span>Sign in with GitHub</span>}
          </Link>
        )}

        {/* Tautan legal, sengaja diletakkan tepat di atas tombol Logout. */}
        {showLabels && (
          <div className={styles.legal}>
            <Link href="/terms" className={styles.legalLink} onClick={() => setMobileOpen(false)}>
              Terms &amp; Conditions
            </Link>
            <Link href="/privacy" className={styles.legalLink} onClick={() => setMobileOpen(false)}>
              Privacy Policy
            </Link>
          </div>
        )}

        {!loading && authed && (
          <button type="button" onClick={handleLogout} disabled={loggingOut} className={styles.item} title={showLabels ? undefined : "Logout"}>
            <svg {...iconProps}>
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
              <path d="M16 17l5-5-5-5" />
              <path d="M21 12H9" />
            </svg>
            {showLabels && <span>{loggingOut ? "Logging out..." : "Logout"}</span>}
          </button>
        )}

        {user && (
          <Link href="/profile" onClick={() => setMobileOpen(false)} className={styles.account} title={showLabels ? undefined : user.username}>
            {user.avatar_url ? (
              <img src={user.avatar_url} alt={user.username} width={24} height={24} loading="lazy" decoding="async" className={styles.avatar} />
            ) : (
              <span className={styles.avatar} style={{ background: "#eaeaea", display: "inline-block" }} />
            )}
            {showLabels && <span className={styles.accountName}>{user.username}</span>}
          </Link>
        )}
      </aside>
    </>
  )
}
