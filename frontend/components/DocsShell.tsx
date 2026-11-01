"use client"

import { usePathname } from "next/navigation"
import { useState, useMemo } from "react"
import Link from "next/link"
import AppShell from "@/components/AppShell"
import { useCurrentUser } from "@/lib/auth"
import ui from "@/app/ui.module.css"
import styles from "@/app/docs/docs.module.css"

const LEFT_SECTIONS = [
  { slug: "", label: "Overview" },
  { slug: "what", label: "What is only4.fun?" },
  { slug: "why", label: "Why it exists" },
  { slug: "vercel", label: "Vercel alternative" },
  { slug: "opensource", label: "Open source" },
  { slug: "how", label: "How it works" },
  { slug: "accounts", label: "Accounts & sign-in" },
  { slug: "deploy", label: "Deploying a site" },
  { slug: "pipeline", label: "Build pipeline" },
  { slug: "frameworks", label: "Supported frameworks" },
  { slug: "subdomain", label: "Project name & subdomain" },
  { slug: "env", label: "Environment variables" },
  { slug: "database", label: "Database with Moogo" },
  { slug: "status", label: "Deployment statuses" },
  { slug: "detail", label: "Project detail page" },
  { slug: "visits", label: "Visit analytics" },
  { slug: "manage", label: "Redeploy & delete" },
  { slug: "domains", label: "Domains & SSL" },
  { slug: "ssr", label: "SSR support" },
  { slug: "api", label: "API reference" },
  { slug: "security", label: "Security & data" },
  { slug: "limits", label: "Limits & scope" },
  { slug: "dos", label: "Do's" },
  { slug: "donts", label: "Don'ts" },
  { slug: "contributing", label: "Contributing" },
  { slug: "troubleshooting", label: "Troubleshooting" },
] as const

export default function DocsShell({ children }: { children: React.ReactNode }) {
  const { user, loaded } = useCurrentUser()
  const pathname = usePathname() || ""
  const currentSlug = pathname.replace(/^\/docs\/?/, "").replace(/\/$/, "") || ""
  const [search, setSearch] = useState("")

  const filtered = useMemo(() => {
    if (!search.trim()) return LEFT_SECTIONS
    const q = search.toLowerCase()
    return LEFT_SECTIONS.filter((s) => s.label.toLowerCase().includes(q))
  }, [search])

  const navContent = (
    <>
      <div className={styles.sidebarTitle}>Documentation</div>
      <div style={{ position: "relative", marginBottom: 12 }}>
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="#8a8a8a"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          style={{ position: "absolute", left: 10, top: "50%", transform: "translateY(-50%)" }}
        >
          <circle cx="11" cy="11" r="8" />
          <line x1="21" y1="21" x2="16.65" y2="16.65" />
        </svg>
        <input
          type="text"
          placeholder="Search..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          style={{
            width: "100%",
            padding: "7px 10px 7px 30px",
            borderRadius: 6,
            border: "1px solid #e2e2e2",
            fontSize: "10pt",
            color: "#171717",
            background: "white",
            outline: "none",
            boxSizing: "border-box",
          }}
        />
        {search && filtered.length === 0 && (
          <div style={{ marginTop: 6, fontSize: "9.5pt", color: "#8a8a8a" }}>
            No results for &quot;{search}&quot;
          </div>
        )}
      </div>
      <nav className={styles.sidebarList}>
        {filtered.map((s) => (
          <Link
            key={s.slug || "overview"}
            href={s.slug ? `/docs/${s.slug}` : "/docs"}
            className={`${styles.sidebarItem} ${currentSlug === s.slug ? styles.sidebarItemActive : ""}`}
          >
            {s.label}
          </Link>
        ))}
      </nav>
    </>
  )

  return (
    <AppShell user={user} loading={!loaded} publicShell>
      <main className={`${ui.page} ${ui.pageWide}`}>
        <div className={ui.topbar}>
          <Link href="/dashboard" className={ui.back}>
            ← Dashboard
          </Link>
          <span className={ui.eyebrow}>Documentation</span>
        </div>
        <div className={styles.layout}>
          <aside className={styles.sidebar}>{navContent}</aside>
          {/* Di layar <=1280px sidebar disembunyikan, jadi daftar isi docs
              tersedia lewat panel lipat ini. key memastikan panel tertutup
              lagi setiap pindah halaman. */}
          <details className={styles.mobileNav} key={currentSlug || "overview"}>
            <summary className={styles.mobileNavSummary}>Browse documentation</summary>
            <div className={styles.mobileNavPanel}>{navContent}</div>
          </details>
          <div className={styles.content}>{children}</div>
        </div>
      </main>
    </AppShell>
  )
}
