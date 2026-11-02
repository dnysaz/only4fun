"use client"

import { useCallback, useEffect, useState } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import AppShell from "@/components/AppShell"
import AuthGate from "@/components/AuthGate"
import ConfirmModal from "@/components/ConfirmModal"
import { apiFetch } from "@/lib/api"
import { useCurrentUser } from "@/lib/auth"
import ui from "../ui.module.css"

interface Project {
  id: string
  repo_url: string
  subdomain: string
  status: string
  views?: number
  env_keys?: string[]
}

function statusMeta(status: string): { color: string; label: string } {
  const map: Record<string, { color: string; label: string }> = {
    deployed: { color: "#1e7e34", label: "Live" },
    building: { color: "#b26a00", label: "Building" },
    pending: { color: "#b26a00", label: "Queued" },
    failed: { color: "#c5221f", label: "Error" },
  }
  return map[status] || { color: "#8a8a8a", label: status }
}

function repoLabel(repoURL: string) {
  if (repoURL.startsWith("zip://")) return `ZIP · ${repoURL.replace("zip://", "")}`
  return repoURL.replace(/^https?:\/\/(www\.)?github\.com\//, "")
}

export default function DashboardPage() {
  const router = useRouter()
  // Sesi dibaca dari cache bersama; AuthGate yang menjaga halaman ini.
  const { user } = useCurrentUser()
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [busyIds, setBusyIds] = useState<Set<string>>(new Set())
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null)
  const [search, setSearch] = useState("")
  const [viewMode, setViewMode] = useState<"list" | "grid">(() => {
    if (typeof window !== "undefined") {
      return (localStorage.getItem("dashboard-view-mode") as "list" | "grid") || "list"
    }
    return "list"
  })

  const loadProjects = useCallback(async () => {
    const res = await apiFetch("/api/projects")
    if (res.ok) {
      const data = await res.json()
      setProjects(data.projects || [])
    }
  }, [])

  useEffect(() => {
    const load = async () => {
      try {
        await loadProjects()
      } catch {
        setError("Could not reach the server. Please try again.")
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [loadProjects])

  // Auto-refresh when any project is building/pending.
  useEffect(() => {
    if (projects.some((p) => p.status === "pending" || p.status === "building")) {
      const timer = setInterval(() => {
        loadProjects()
      }, 3000)
      return () => clearInterval(timer)
    }
  }, [projects, loadProjects])

  const handleRedeploy = async (id: string) => {
    setBusyIds((prev) => new Set(prev).add(id))
    setError("")
    try {
      const res = await apiFetch(`/api/projects/${id}/deploy`, { method: "POST" })
      if (!res.ok) {
        const txt = await res.text()
        let msg = txt
        try {
          msg = JSON.parse(txt).error || txt
        } catch {}
        setError(msg || "Redeploy failed")
        return
      }
      await loadProjects()
    } finally {
      setBusyIds((prev) => {
        const next = new Set(prev)
        next.delete(id)
        return next
      })
    }
  }

  const performDelete = async (id: string) => {
    setBusyIds((prev) => new Set(prev).add(id))
    setError("")
    try {
      const res = await apiFetch(`/api/projects/${id}`, { method: "DELETE" })
      if (!res.ok) {
        setError("Could not delete project")
        return
      }
      await loadProjects()
    } finally {
      setBusyIds((prev) => {
        const next = new Set(prev)
        next.delete(id)
        return next
      })
    }
  }

  const filtered = projects.filter((p) => {
    if (!search) return true
    const q = search.toLowerCase()
    return p.subdomain.toLowerCase().includes(q) || p.repo_url.toLowerCase().includes(q)
  })

  const liveCount = projects.filter((p) => p.status === "deployed").length
  const totalViews = projects.reduce((sum, p) => sum + (p.views || 0), 0)

  return (
    <AuthGate>
      <AppShell user={user}>
      <main className={`${ui.page} ${ui.pageWide}`}>
        <div className={ui.topbar}>
          <span className={ui.eyebrow}>Dashboard</span>
          <Link href="/projects/new" className={`${ui.btn} ${ui.btnPrimary}`}>
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
              <line x1="12" y1="5" x2="12" y2="19" />
              <line x1="5" y1="12" x2="19" y2="12" />
            </svg>
            New project
          </Link>
        </div>

        <header className={ui.head}>
          <div className={ui.headMain}>
            <h1 className={ui.title}>Projects.</h1>
            <p className={ui.lede}>
              Everything you have shipped, in one place. Open a project to see its builds, traffic, domain, and
              environment variables.
            </p>
          </div>
        </header>

        <div className={ui.stats}>
          <div className={ui.stat}>
            <div className={ui.statLabel}>Projects</div>
            <div className={ui.statValue}>{projects.length}</div>
            <div className={ui.statHint}>total</div>
          </div>
          <div className={ui.stat}>
            <div className={ui.statLabel}>Live</div>
            <div className={ui.statValue}>{liveCount}</div>
            <div className={ui.statHint}>currently deployed</div>
          </div>
          <div className={ui.stat}>
            <div className={ui.statLabel}>Page views</div>
            <div className={ui.statValue}>{totalViews.toLocaleString("en-GB")}</div>
            <div className={ui.statHint}>all projects</div>
          </div>
        </div>

        <ConfirmModal
          open={Boolean(deleteTarget)}
          title="Delete project"
          message="Delete this project permanently? The subdomain, published files, releases, and DNS record are removed. The GitHub repository is not touched. This cannot be undone."
          onConfirm={() => {
            const target = deleteTarget
            setDeleteTarget(null)
            if (target) performDelete(target)
          }}
          onCancel={() => setDeleteTarget(null)}
        />

        <section className={ui.section}>
          <div className={ui.sectionHead}>
            <div className={ui.sectionHeadMain}>
              <h2 className={ui.sectionTitle}>All projects</h2>
              <span className={ui.sectionCount}>{filtered.length}</span>
            </div>
            <div className={ui.viewToggle}>
              <button
                className={`${ui.viewToggleBtn} ${viewMode === "list" ? ui.viewToggleBtnActive : ""}`}
                onClick={() => {
                  setViewMode("list")
                  localStorage.setItem("dashboard-view-mode", "list")
                }}
                aria-label="List view"
              >
                List
              </button>
              <button
                className={`${ui.viewToggleBtn} ${viewMode === "grid" ? ui.viewToggleBtnActive : ""}`}
                onClick={() => {
                  setViewMode("grid")
                  localStorage.setItem("dashboard-view-mode", "grid")
                }}
                aria-label="Grid view"
              >
                Grid
              </button>
            </div>
          </div>

          {projects.length > 0 && (
            <div className={ui.searchWrap}>
              <input
                type="text"
                className={ui.search}
                placeholder="Search by name or repository"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                aria-label="Search projects"
              />
              <div className={ui.searchMeta}>
                <span>
                  {filtered.length} of {projects.length} projects
                </span>
                <span>Status refreshes on reload</span>
              </div>
            </div>
          )}

          {error && (
            <p className={ui.errorText} role="alert">
              <span className={ui.dot} aria-hidden="true" />
              {error}
            </p>
          )}

          {loading ? (
            <div style={{ marginTop: 22, display: "grid", gap: 26 }}>
              {[1, 2, 3].map((i) => (
                <div key={i} style={{ display: "flex", alignItems: "center", gap: 24 }}>
                  <div style={{ flex: 1, display: "grid", gap: 8 }}>
                    <div className={ui.skeleton} style={{ width: "38%" }} />
                    <div className={ui.skeleton} style={{ width: "62%", opacity: 0.6 }} />
                  </div>
                  <div className={ui.skeleton} style={{ width: 72 }} />
                  <div className={ui.skeleton} style={{ width: 96 }} />
                </div>
              ))}
            </div>
          ) : projects.length === 0 ? (
            <div style={{ padding: "30px 0 10px" }}>
              <p className={ui.empty}>
                No projects yet. Deploy from a GitHub repository or drop a ZIP archive to get your first URL.
              </p>
              <Link href="/projects/new" className={`${ui.btn} ${ui.btnPrimary}`} style={{ marginTop: 6 }}>
                Create your first project
              </Link>
            </div>
          ) : filtered.length === 0 ? (
            <p className={ui.empty}>No projects match &quot;{search}&quot;.</p>
          ) : viewMode === "grid" ? (
            <div className={ui.grid}>
              {filtered.map((p) => {
                const st = statusMeta(p.status)
                const isZip = p.repo_url.startsWith("zip://")
                return (
                  <div
                    key={p.id}
                    className={ui.gridCard}
                    onClick={() => router.push(`/projects/${p.id}`)}
                  >
                    <h3 className={ui.gridCardTitle}>{p.subdomain}</h3>
                    <p className={ui.gridCardMeta}>{repoLabel(p.repo_url)}</p>
                    <div className={ui.gridCardFooter}>
                      <span className={ui.status} style={{ color: st.color }}>
                        <span className={ui.statusDot} aria-hidden="true" />
                        {st.label}
                      </span>
                      <span style={{ fontSize: "12px", color: "#8a8a8a" }}>
                        {(p.views ?? 0).toLocaleString("en-GB")} views
                      </span>
                    </div>
                    <div style={{ marginTop: 12, display: "flex", gap: 6, flexWrap: "wrap" }}>
                      <a
                        href={`https://${p.subdomain}.only4.fun`}
                        target="_blank"
                        rel="noreferrer"
                        onClick={(e) => e.stopPropagation()}
                        className={`${ui.btn} ${ui.btnTiny}`}
                      >
                        Visit ↗
                      </a>
                      {!isZip && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation()
                            handleRedeploy(p.id)
                          }}
                          disabled={busyIds.has(p.id)}
                          className={`${ui.btn} ${ui.btnTiny}`}
                        >
                          {busyIds.has(p.id) ? "Working..." : "Redeploy"}
                        </button>
                      )}
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          setDeleteTarget(p.id)
                        }}
                        disabled={busyIds.has(p.id)}
                        className={`${ui.btn} ${ui.btnTiny} ${ui.btnDanger}`}
                      >
                        Delete
                      </button>
                    </div>
                  </div>
                )
              })}
            </div>
          ) : (
            <div className={ui.list}>
              {filtered.map((p) => {
                const st = statusMeta(p.status)
                const isZip = p.repo_url.startsWith("zip://")
                return (
                  <div
                    key={p.id}
                    className={`${ui.listRow} ${ui.listRowClickable}`}
                    onClick={() => router.push(`/projects/${p.id}`)}
                  >
                    <div className={ui.rowMain}>
                      <Link
                        href={`/projects/${p.id}`}
                        className={ui.rowTitle}
                        onClick={(e) => e.stopPropagation()}
                      >
                        {p.subdomain}
                      </Link>
                      <div className={ui.rowMeta}>{repoLabel(p.repo_url)}</div>
                    </div>

                    <div className={ui.rowSide}>
                      <span className={ui.status} style={{ color: st.color }}>
                        <span className={ui.statusDot} aria-hidden="true" />
                        {st.label}
                      </span>

                      <div className={ui.metric}>
                        <div className={ui.metricValue}>{(p.views ?? 0).toLocaleString("en-GB")}</div>
                        <div className={ui.metricLabel}>views</div>
                      </div>

                      <div className={ui.rowActions}>
                        <a
                          href={`https://${p.subdomain}.only4.fun`}
                          target="_blank"
                          rel="noreferrer"
                          onClick={(e) => e.stopPropagation()}
                          className={`${ui.btn} ${ui.btnTiny}`}
                        >
                          Visit ↗
                        </a>
                        {!isZip && (
                          <button
                            onClick={(e) => {
                              e.stopPropagation()
                              handleRedeploy(p.id)
                            }}
                            disabled={busyIds.has(p.id)}
                            className={`${ui.btn} ${ui.btnTiny}`}
                          >
                            {busyIds.has(p.id) ? "Working..." : "Redeploy"}
                          </button>
                        )}
                        <button
                          onClick={(e) => {
                            e.stopPropagation()
                            setDeleteTarget(p.id)
                          }}
                          disabled={busyIds.has(p.id)}
                          className={`${ui.btn} ${ui.btnTiny} ${ui.btnDanger}`}
                        >
                          Delete
                        </button>
                      </div>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </section>
      </main>
      </AppShell>
    </AuthGate>
  )
}
