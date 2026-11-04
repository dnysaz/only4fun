"use client"

import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import Link from "next/link"
import { useParams, useRouter } from "next/navigation"
import AppShell from "@/components/AppShell"
import AuthGate from "@/components/AuthGate"
import BuildTerminal from "@/components/BuildTerminal"
import ConfirmModal from "@/components/ConfirmModal"
import SidePanel from "@/components/SidePanel"
import VisitsChart, { buildVisitSeries, formatDayLabel } from "@/components/VisitsChart"
import { apiFetch } from "@/lib/api"
import { safeExternalUrl } from "@/lib/url"
import { useCurrentUser } from "@/lib/auth"
import ui from "../../ui.module.css"

interface DailyVisit {
  day: string
  count: number
}

interface Project {
  id: string
  repo_url: string
  subdomain: string
  status: string
  created_at: string
  updated_at: string
  last_visited_at?: string
  views: number
  env_keys: string[] | null
  custom_domain?: string
  visits?: DailyVisit[]
}

interface RepoInfo {
  full_name: string
  description: string
  html_url: string
  private: boolean
  language: string
  default_branch: string
  stars: number
  forks: number
  open_issues: number
  pushed_at: string
  updated_at: string
}

interface WebhookInfo {
  url: string
  secret: string
  events: string[]
}

interface Deployment {
  id: string
  status: string
  log: string
  release?: string
  created_at: string
  updated_at: string
}

function statusMeta(status: string): { color: string; label: string } {
  const map: Record<string, { color: string; label: string }> = {
    deployed: { color: "#1e7e34", label: "Live" },
    building: { color: "#b26a00", label: "Building" },
    pending: { color: "#b26a00", label: "Queued" },
    failed: { color: "#c5221f", label: "Failed" },
  }
  return map[status] || { color: "#8a8a8a", label: status }
}

function fmtDate(value?: string) {
  if (!value) return "—"
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return "—"
  return d.toLocaleString("en-GB", { dateStyle: "medium", timeStyle: "short" })
}

function repoLabel(repoURL: string) {
  if (repoURL.startsWith("zip://")) return `ZIP archive · ${repoURL.replace("zip://", "")}`
  return repoURL.replace(/^https?:\/\/(www\.)?github\.com\//, "")
}

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className={ui.kvRow}>
      <span className={ui.kvLabel}>{label}</span>
      <span className={ui.kvValue}>{children}</span>
    </div>
  )
}

function Stat({ label, value, hint }: { label: string; value: React.ReactNode; hint?: string }) {
  return (
    <div className={ui.stat}>
      <div className={ui.statLabel}>{label}</div>
      <div className={ui.statValue}>{value}</div>
      {hint && <div className={ui.statHint}>{hint}</div>}
    </div>
  )
}

export default function ProjectDetailPage() {
  const params = useParams<{ id: string }>()
  const router = useRouter()
  const id = params?.id

  // Sesi diambil dari cache bersama; AuthGate yang menjaga halaman ini.
  const { user } = useCurrentUser()
  const [project, setProject] = useState<Project | null>(null)
  const [repo, setRepo] = useState<RepoInfo | null>(null)
  const [repoError, setRepoError] = useState("")
  const [deployments, setDeployments] = useState<Deployment[]>([])
  const [webhook, setWebhook] = useState<WebhookInfo | null>(null)
  const [webhookOpen, setWebhookOpen] = useState(false)
  const [copiedField, setCopiedField] = useState("")
  const [domainInput, setDomainInput] = useState("")
  const [domainSaving, setDomainSaving] = useState(false)
  const [domainMessage, setDomainMessage] = useState("")
  const [envVars, setEnvVars] = useState<Record<string, string>>({})
  const [envDraft, setEnvDraft] = useState("")
  const [editingEnv, setEditingEnv] = useState(false)
  const [envOpen, setEnvOpen] = useState(false)
  const [domainOpen, setDomainOpen] = useState(false)
  const [openLog, setOpenLog] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [notice, setNotice] = useState("")
  const [busy, setBusy] = useState(false)
  const [clearModalOpen, setClearModalOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  // Menandai redeploy yang dipicu dari halaman ini supaya notice bisa berubah
  // dari "Redeploy started." menjadi "Redeploy finished." saat build selesai.
  const deployAwaiting = useRef(false)
  const previousDeploymentID = useRef<string | null>(null)

  const liveURL = useMemo(() => {
    if (!project) return ""
    return `https://${project.subdomain}.only4.fun`
  }, [project])

  // repo_url berasal dari database; jangan pernah jadikan href mentah-mentah.
  const repoURL = useMemo(() => safeExternalUrl(project?.repo_url), [project])

  // Project dari unggahan ZIP tidak punya repo untuk di-redeploy.
  const isZipProject = Boolean(project?.repo_url?.startsWith("zip://"))

  useEffect(() => {
    setDomainInput(project?.custom_domain || "")
  }, [project?.custom_domain])

  const loadProject = useCallback(async () => {
    if (!id) return
    const res = await apiFetch(`/api/projects/${id}`)
    if (res.status === 401) {
      router.push("/")
      return
    }
    if (!res.ok) {
      const txt = await res.text()
      let msg = txt
      try {
        msg = JSON.parse(txt).error || txt
      } catch {}
      setError(msg || `Could not load project (${res.status})`)
      return
    }
    const data = await res.json()
    setProject(data.project)
  }, [id, router])

  const loadDeployments = useCallback(async () => {
    if (!id) return
    const res = await apiFetch(`/api/projects/${id}/deployments`)
    if (res.ok) {
      const data = await res.json()
      setDeployments(data.deployments || [])
    }
  }, [id])

  const loadWebhook = useCallback(async () => {
    if (!id) return
    const res = await apiFetch(`/api/projects/${id}/webhook`)
    if (res.ok) {
      const data = await res.json()
      setWebhook(data.webhook || null)
    }
  }, [id])

  const copyToClipboard = async (field: string, value: string) => {
    try {
      await navigator.clipboard.writeText(value)
      setCopiedField(field)
      setTimeout(() => setCopiedField(""), 1500)
    } catch {
      // Clipboard bisa diblokir browser; nilai tetap terlihat untuk disalin manual.
    }
  }

  const loadEnv = useCallback(async () => {
    if (!id) return
    const res = await apiFetch(`/api/projects/${id}/env`)
    if (res.ok) {
      const data = await res.json()
      // Hanya key yang dikembalikan (nilai tidak pernah dikirim ke browser).
      const keys: string[] = data.env_keys || []
      const vars: Record<string, string> = {}
      keys.forEach((k) => {
        vars[k] = ""
      })
      setEnvVars(vars)
      setEnvDraft("")
    }
  }, [id])

  useEffect(() => {
    const bootstrap = async () => {
      await loadProject()
      await Promise.all([loadDeployments(), loadEnv(), loadWebhook()])

      if (id) {
        const repoRes = await apiFetch(`/api/projects/${id}/repo`)
        if (repoRes.ok) {
          const repoData = await repoRes.json()
          setRepo(repoData.repo)
        } else {
          const txt = await repoRes.text()
          let msg = txt
          try {
            msg = JSON.parse(txt).error || txt
          } catch {}
          setRepoError(msg || "Repository details are unavailable")
        }
      }
      setLoading(false)
    }
    bootstrap()
  }, [id, loadProject, loadDeployments, loadEnv, loadWebhook])

  // Poll status selama masih diproses.
  useEffect(() => {
    if (!project) return
    if (project.status !== "pending" && project.status !== "building") return
    const timer = setInterval(async () => {
      await loadProject()
      await loadDeployments()
    }, 2000)
    return () => clearInterval(timer)
  }, [project, loadProject, loadDeployments])

  // Perbarui notice ketika deployment baru yang dipicu dari halaman ini selesai.
  useEffect(() => {
    if (!deployAwaiting.current) return
    const latest = deployments[0]
    if (!latest || latest.id === previousDeploymentID.current) return
    if (latest.status !== "success" && latest.status !== "failed") return

    deployAwaiting.current = false
    if (latest.status === "success") {
      setNotice("Redeploy finished.")
    } else {
      setNotice("")
      setError("Redeploy failed.")
    }
  }, [deployments])

  const handleRollback = async (deployment: Deployment) => {
    if (busy || !id) return
    setBusy(true)
    setError("")
    setNotice("")
    try {
      const res = await apiFetch(`/api/projects/${id}/deployments/${deployment.id}/rollback`, { method: "POST" })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) {
        setError(data.error || "Could not roll back to this deployment.")
        return
      }
      setNotice("Rolled back to the selected deployment.")
      await Promise.all([loadProject(), loadDeployments()])
    } finally {
      setBusy(false)
    }
  }

  const handleRedeploy = async () => {
    if (!id) return
    setBusy(true)
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
      deployAwaiting.current = true
      previousDeploymentID.current = deployments[0]?.id || null
      setNotice("Redeploy started.")
      await loadProject()
      await loadDeployments()
    } finally {
      setBusy(false)
    }
  }

  const performDelete = async () => {
    if (!id) return
    setBusy(true)
    try {
      const res = await apiFetch(`/api/projects/${id}`, { method: "DELETE" })
      if (res.ok) {
        router.push("/dashboard")
        return
      }
      setError("Could not delete project")
    } finally {
      setBusy(false)
    }
  }

  const handleSaveEnv = async () => {
    if (!id) return
    const parsed: Record<string, string> = {}
    envDraft.split("\n").forEach((line) => {
      if (!line.trim()) return
      const idx = line.indexOf("=")
      if (idx <= 0) return
      parsed[line.slice(0, idx).trim()] = line.slice(idx + 1).trim()
    })
    setBusy(true)
    setError("")
    try {
      const res = await apiFetch(`/api/projects/${id}/env`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ env_vars: parsed }),
      })
      if (!res.ok) {
        const txt = await res.text()
        let msg = txt
        try {
          msg = JSON.parse(txt).error || txt
        } catch {}
        setError(msg || "Could not save environment variables")
        return
      }
      setEnvVars(parsed)
      setEditingEnv(false)
      setNotice("Environment variables saved. Redeploy so the new values are used in the next build.")
      await loadProject()
      await loadEnv()
    } finally {
      setBusy(false)
    }
  }

  const saveDomain = async (domain: string) => {
    if (!id || domainSaving) return
    setDomainSaving(true)
    setDomainMessage("")
    try {
      const res = await apiFetch(`/api/projects/${id}/domain`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ domain }),
      })
      const data = await res.json().catch(() => ({}))
      if (!res.ok) {
        setDomainMessage(data.error || "Could not save the domain.")
        return
      }
      setDomainMessage(data.warning || (domain ? "Domain saved." : "Domain removed."))
      await loadProject()
    } finally {
      setDomainSaving(false)
    }
  }

  const visits = project?.visits || []
  const visitSeries = useMemo(() => buildVisitSeries(visits), [visits])
  const last14 = useMemo(() => visitSeries.reduce((sum, v) => sum + v.count, 0), [visitSeries])
  const peak = useMemo(() => Math.max(0, ...visitSeries.map((v) => v.count)), [visitSeries])
  const peakDay = useMemo(() => (peak > 0 ? visitSeries.find((v) => v.count === peak) : undefined), [visitSeries, peak])
  const dailyAverage = (last14 / (visitSeries.length || 1)).toFixed(1)
  // Only the deployment currently being built belongs in the terminal; older
  // logs must not flash while a fresh redeploy is still queued.
  const activeDeployment = deployments.find((d) => d.status === "building") || null

  if (loading) {
    return (
      <AuthGate>
        <AppShell user={user}>
        <main className={ui.page}>
          <div style={{ display: "grid", gap: 12 }}>
            <div className={ui.skeleton} style={{ width: 180 }} />
            <div className={ui.skeleton} style={{ width: 280, height: 28 }} />
            <div className={ui.skeleton} style={{ width: "55%", marginTop: 12 }} />
          </div>
        </main>
        </AppShell>
      </AuthGate>
    )
  }

  if (!project) {
    return (
      <AuthGate>
        <AppShell user={user}>
        <main className={ui.page}>
          <Link href="/dashboard" className={ui.back}>
            ← Dashboard
          </Link>
          <p className={ui.errorText} style={{ marginTop: 20 }} role="alert">
            <span className={ui.dot} aria-hidden="true" />
            {error || "Project not found."}
          </p>
        </main>
        </AppShell>
      </AuthGate>
    )
  }

  const projectStatus = statusMeta(project.status)
  const envKeys = Object.keys(envVars)

  return (
    <AuthGate>
      <AppShell user={user}>
      <main className={`${ui.page} ${ui.pageWide}`}>
        <div className={ui.topbar}>
          <Link href="/dashboard" className={ui.back}>
            ← Dashboard
          </Link>
          <span className={ui.eyebrow}>Project</span>
        </div>

        <header className={ui.head}>
          <div className={ui.headMain}>
            <h1 className={ui.title}>{project.subdomain}</h1>
            <div style={{ display: "flex", alignItems: "center", gap: 14, flexWrap: "wrap" }}>
              <span className={ui.status} style={{ color: projectStatus.color }}>
                <span className={ui.statusDot} aria-hidden="true" />
                {projectStatus.label}
              </span>
              <a href={liveURL} target="_blank" rel="noreferrer" style={{ fontSize: 13.5, color: "#757575" }}>
                {project.subdomain}.only4.fun ↗
              </a>
            </div>
          </div>
          <div className={ui.actions}>
            {!isZipProject && (
              <button className={`${ui.btn} ${ui.btnPrimary}`} onClick={handleRedeploy} disabled={busy}>
                Redeploy
              </button>
            )}
            <button className={`${ui.btn} ${ui.btnGhost}`} onClick={() => setClearModalOpen(true)} disabled={busy}>
              Clear history
            </button>
            <button className={`${ui.btn} ${ui.btnDanger}`} onClick={() => setDeleteOpen(true)} disabled={busy}>
              Delete
            </button>
          </div>
        </header>

        <ConfirmModal
          open={clearModalOpen}
          title="Clear Deployments"
          message="Permanently delete every deployment log for this project? The deployed site and its files are not affected, but the build history cannot be recovered."
          onConfirm={() => {
            apiFetch(`/api/projects/${id}/deployments`, { method: "DELETE" }).finally(() => {
              setDeployments([])
              setClearModalOpen(false)
            })
          }}
          onCancel={() => setClearModalOpen(false)}
        />

        <ConfirmModal
          open={deleteOpen}
          title="Delete project"
          message="Delete this project permanently? The subdomain, published files, releases, and DNS record are removed. The GitHub repository is not touched. This cannot be undone."
          onConfirm={() => {
            setDeleteOpen(false)
            performDelete()
          }}
          onCancel={() => setDeleteOpen(false)}
        />

        {error ? (
          <p className={ui.errorText} role="alert">
            <span className={ui.dot} aria-hidden="true" />
            {error}
          </p>
        ) : notice ? (
          <p className={ui.noticeText}>
            <span className={ui.dot} aria-hidden="true" />
            {notice}
          </p>
        ) : null}

        {(project.status === "pending" || project.status === "building") && (
          <div style={{ marginTop: 20 }}>
            <BuildTerminal projectStatus={project.status} deployment={activeDeployment} />
          </div>
        )}

        {/* 01 — Overview */}
        <section className={ui.section}>
          <div className={ui.sectionHead}>
            <div className={ui.sectionHeadMain}>
              <h2 className={ui.sectionTitle}>Overview</h2>
            </div>
            <span className={ui.eyebrow}>Details</span>
          </div>

          <div className={ui.sectionBody}>
            <div className={`${ui.kv} ${ui.kvWide}`}>
              <Row label="Repository">
                {isZipProject ? (
                  repoLabel(project.repo_url)
                ) : repoURL ? (
                  <a href={repoURL} target="_blank" rel="noreferrer">
                    {repoLabel(project.repo_url)} ↗
                  </a>
                ) : (
                  repoLabel(project.repo_url)
                )}
              </Row>
              <Row label="Subdomain">{project.subdomain}.only4.fun</Row>
              <Row label="Custom domain">{project.custom_domain || "—"}</Row>
              <Row label="Status">{projectStatus.label}</Row>
              <Row label="Created">{fmtDate(project.created_at)}</Row>
              <Row label="Last updated">{fmtDate(project.updated_at)}</Row>
              {!isZipProject && repo && (
                <>
                  <Row label="Visibility">{repo.private ? "Private" : "Public"}</Row>
                  <Row label="Default branch">{repo.default_branch || "—"}</Row>
                  <Row label="Language">{repo.language || "—"}</Row>
                  <Row label="Stars / forks / issues">
                    {repo.stars} / {repo.forks} / {repo.open_issues}
                  </Row>
                  <Row label="Last push">{fmtDate(repo.pushed_at)}</Row>
                </>
              )}
            </div>

            {!isZipProject && (
              <p className={ui.note}>
                {repoError
                  ? repoError
                  : !repo
                    ? "Loading repository information..."
                    : repo.description || "This repository has no description."}
              </p>
            )}
            {isZipProject && (
              <p className={ui.note}>
                Created from a ZIP upload. Static files only; redeploy a new project to publish a new version.
              </p>
            )}
          </div>
        </section>

        {/* 02 — Traffic */}
        <section className={ui.section}>
          <div className={ui.sectionHead}>
            <div className={ui.sectionHeadMain}>
              <h2 className={ui.sectionTitle}>Traffic</h2>
            </div>
            <span className={ui.eyebrow}>Last 14 days</span>
          </div>

          <div className={ui.stats}>
            <Stat label="Total views" value={(project.views ?? 0).toLocaleString("en-GB")} hint="all time" />
            <Stat label="Last 14 days" value={last14.toLocaleString("en-GB")} hint="page views" />
            <Stat label="Daily average" value={dailyAverage} hint="views / day" />
            <Stat label="Peak day" value={peak} hint={peakDay ? formatDayLabel(peakDay.day) : "no visits yet"} />
          </div>

          <div style={{ marginTop: 24 }}>
            {visits.length === 0 ? (
              <p className={ui.empty}>No visits recorded yet.</p>
            ) : (
              <VisitsChart visits={visits} />
            )}
          </div>
          <p className={ui.note}>Last visit: {fmtDate(project.last_visited_at)}</p>
        </section>

        {/* 03 — Deployments */}
        <section className={ui.section}>
          <div className={ui.sectionHead}>
            <div className={ui.sectionHeadMain}>
              <h2 className={ui.sectionTitle}>Deployments</h2>
              <span className={ui.sectionCount}>{deployments.length}</span>
            </div>
          </div>

          <div className={ui.sectionBody}>
            {deployments.length === 0 ? (
              <p className={ui.empty}>No deployments yet.</p>
            ) : (
              <div className={ui.list}>
                {deployments.map((d) => {
                  const st = statusMeta(d.status === "success" ? "deployed" : d.status === "failed" ? "failed" : d.status)
                  return (
                    <div key={d.id}>
                      <div className={ui.listRow} style={{ borderBottom: openLog === d.id ? "none" : undefined }}>
                        <div className={ui.rowMain}>
                          <div style={{ display: "flex", alignItems: "center", gap: 10, flexWrap: "wrap" }}>
                            <span className={ui.status} style={{ color: st.color }}>
                              <span className={ui.statusDot} aria-hidden="true" />
                              {st.label}
                            </span>
                            <span className={`${ui.mono}`} style={{ fontSize: 12.5, color: "#a0a0a0" }}>
                              #{d.id}
                            </span>
                          </div>
                          <div className={ui.rowMeta}>{fmtDate(d.created_at)}</div>
                        </div>
                        <div className={ui.rowActions}>
                          {d.status === "success" && d.release && (
                            <button
                              onClick={() => handleRollback(d)}
                              disabled={busy}
                              title={`Make ${d.release} the active version`}
                              className={`${ui.btn} ${ui.btnTiny} ${ui.btnGhost}`}
                            >
                              Rollback
                            </button>
                          )}
                          <button
                            onClick={() => setOpenLog(openLog === d.id ? null : d.id)}
                            className={`${ui.btn} ${ui.btnTiny} ${ui.btnGhost}`}
                          >
                            {openLog === d.id ? "Hide log" : "View log"}
                          </button>
                        </div>
                      </div>
                      {openLog === d.id && <pre className={ui.log}>{d.log || "(empty log)"}</pre>}
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        </section>

        {/* 04 — Environment */}
        <section className={ui.section}>
          <div className={ui.sectionHead}>
            <div className={ui.sectionHeadMain}>
              <h2 className={ui.sectionTitle}>Environment variables</h2>
              <span className={ui.sectionCount}>{envKeys.length}</span>
            </div>
          </div>
          <p className={ui.sectionHint}>
            Keys and secrets injected into the build. Values stay encrypted and are never sent to the browser.
          </p>

          <div className={ui.sectionBody}>
            <div className={ui.list}>
              <div className={ui.listRow}>
                <div className={ui.rowMain}>
                  <span className={ui.rowTitle}>
                    {envKeys.length === 0
                      ? "No variables yet"
                      : `${envKeys.length} variable${envKeys.length === 1 ? "" : "s"}`}
                  </span>
                  <div className={ui.rowMeta}>Encrypted at rest · applied on the next build</div>
                </div>
                <div className={ui.rowActions}>
                  <button
                    className={`${ui.btn} ${ui.btnSmall} ${ui.btnGhost}`}
                    onClick={() => setEnvOpen(true)}
                  >
                    Configure
                  </button>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* 05 — Domain */}
        <section className={ui.section}>
          <div className={ui.sectionHead}>
            <div className={ui.sectionHeadMain}>
              <h2 className={ui.sectionTitle}>Custom domain</h2>
            </div>
            {project.custom_domain && (
              <a href={`https://${project.custom_domain}`} target="_blank" rel="noreferrer" className={ui.eyebrow}>
                {project.custom_domain} ↗
              </a>
            )}
          </div>
          <p className={ui.sectionHint}>Serve this project on a domain you own, with SSL issued automatically.</p>

          <div className={ui.sectionBody}>
            <div className={ui.list}>
              <div className={ui.listRow}>
                <div className={ui.rowMain}>
                  <span className={ui.rowTitle}>{project.custom_domain || "No custom domain"}</span>
                  <div className={ui.rowMeta}>
                    CNAME to {project.subdomain}.only4.fun
                  </div>
                </div>
                <div className={ui.rowActions}>
                  <button
                    className={`${ui.btn} ${ui.btnSmall} ${ui.btnGhost}`}
                    onClick={() => setDomainOpen(true)}
                  >
                    Configure
                  </button>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* 06 — Webhook */}
        <section className={ui.section}>
          <div className={ui.sectionHead}>
            <div className={ui.sectionHeadMain}>
              <h2 className={ui.sectionTitle}>Auto-deploy webhook</h2>
            </div>
            <span className={ui.eyebrow}>GitHub</span>
          </div>
          <p className={ui.sectionHint}>
            Connect a GitHub webhook so every push to the default branch deploys automatically.
          </p>

          <div className={ui.sectionBody}>
            <div className={ui.list}>
              <div className={ui.listRow}>
                <div className={ui.rowMain}>
                  <span className={ui.rowTitle}>GitHub push webhook</span>
                  <div className={ui.rowMeta}>
                    {webhook ? webhook.url : "Webhook details are unavailable."}
                  </div>
                </div>
                <div className={ui.rowActions}>
                  <button
                    className={`${ui.btn} ${ui.btnSmall} ${ui.btnGhost}`}
                    onClick={() => setWebhookOpen(true)}
                    disabled={!webhook}
                  >
                    Configure
                  </button>
                </div>
              </div>
            </div>
          </div>
        </section>

        <SidePanel
          open={envOpen}
          title="Environment variables"
          subtitle="Injected at build time · encrypted at rest"
          onClose={() => {
            setEnvOpen(false)
            setEditingEnv(false)
          }}
        >
          {error && (
            <p className={ui.errorText} style={{ marginTop: 0, marginBottom: 16 }} role="alert">
              <span className={ui.dot} aria-hidden="true" />
              {error}
            </p>
          )}

          {editingEnv ? (
            <>
              <label className={ui.fieldLabel} htmlFor="env-draft">
                Variables
              </label>
              <textarea
                id="env-draft"
                value={envDraft}
                onChange={(e) => setEnvDraft(e.target.value)}
                rows={9}
                className={ui.textarea}
                placeholder={"API_KEY=secret\nNODE_ENV=production"}
              />
              <p className={ui.note}>
                One KEY=value per line. An empty value removes the key. Values are encrypted at rest and never sent to
                the browser.
              </p>
              <div style={{ display: "flex", gap: 8, flexWrap: "wrap", marginTop: 16 }}>
                <button
                  onClick={handleSaveEnv}
                  disabled={busy}
                  className={`${ui.btn} ${ui.btnPrimary} ${ui.btnSmall}`}
                >
                  {busy ? "Saving..." : "Save variables"}
                </button>
                <button
                  onClick={() => setEditingEnv(false)}
                  disabled={busy}
                  className={`${ui.btn} ${ui.btnSmall} ${ui.btnGhost}`}
                >
                  Cancel
                </button>
              </div>
            </>
          ) : envKeys.length === 0 ? (
            <>
              <p className={ui.empty} style={{ marginTop: 0 }}>No environment variables yet.</p>
              <button
                className={`${ui.btn} ${ui.btnPrimary} ${ui.btnSmall}`}
                onClick={() => setEditingEnv(true)}
              >
                Add variables
              </button>
            </>
          ) : (
            <>
              <div className={ui.kv}>
                {envKeys.map((k) => (
                  <div key={k} className={ui.kvRow}>
                    <code className={`${ui.mono}`} style={{ fontSize: 13, fontWeight: 600 }}>
                      {k}
                    </code>
                    <span
                      className={ui.kvValue}
                      style={{ display: "inline-flex", alignItems: "center", gap: 5, fontSize: 12 }}
                    >
                      <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                        <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                        <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                      </svg>
                      locked
                    </span>
                  </div>
                ))}
              </div>
              <p className={ui.note}>
                Values are encrypted at rest and never sent to the browser. Update them here, then redeploy so the next
                build uses the new values.
              </p>
              <button
                className={`${ui.btn} ${ui.btnSmall} ${ui.btnGhost}`}
                onClick={() => setEditingEnv(true)}
                style={{ marginTop: 16 }}
              >
                Edit variables
              </button>
            </>
          )}
        </SidePanel>

        <SidePanel
          open={domainOpen}
          title="Custom domain"
          subtitle="Point a domain you own to this project"
          onClose={() => {
            setDomainOpen(false)
            setDomainMessage("")
          }}
        >
          <p className={ui.note} style={{ marginTop: 0 }}>
            Create a <strong>CNAME</strong> record for your domain targeting{" "}
            <code className={ui.code}>{project.subdomain}.only4.fun</code>. SSL is issued once the record resolves.
          </p>

          <div style={{ marginTop: 20 }}>
            <label className={ui.fieldLabel} htmlFor="custom-domain">
              Domain
            </label>
            <input
              id="custom-domain"
              value={domainInput}
              onChange={(e) => setDomainInput(e.target.value)}
              placeholder="www.example.com"
              className={ui.input}
            />
            <div style={{ display: "flex", gap: 8, flexWrap: "wrap", marginTop: 14 }}>
              <button
                onClick={() => saveDomain(domainInput.trim().toLowerCase())}
                disabled={domainSaving}
                className={`${ui.btn} ${ui.btnPrimary} ${ui.btnSmall}`}
              >
                {domainSaving ? "Saving..." : "Save domain"}
              </button>
              {project.custom_domain && (
                <button
                  onClick={() => saveDomain("")}
                  disabled={domainSaving}
                  className={`${ui.btn} ${ui.btnDanger} ${ui.btnSmall}`}
                >
                  Remove
                </button>
              )}
            </div>
          </div>

          {project.custom_domain && (
            <p className={ui.note}>
              Active domain:{" "}
              <a href={`https://${project.custom_domain}`} target="_blank" rel="noreferrer">
                {project.custom_domain} ↗
              </a>
            </p>
          )}

          {domainMessage && (
            <p
              className={
                domainMessage.includes("saved") || domainMessage.includes("removed") ? ui.noticeText : ui.errorText
              }
              role="status"
            >
              <span className={ui.dot} aria-hidden="true" />
              {domainMessage}
            </p>
          )}
        </SidePanel>

        <SidePanel
          open={webhookOpen}
          title="Auto-deploy webhook"
          subtitle="GitHub → Settings → Webhooks → Add webhook"
          onClose={() => setWebhookOpen(false)}
        >
          {webhook ? (
            <>
              <p className={ui.note} style={{ marginTop: 0, marginBottom: 18 }}>
                Use content type <code className={ui.code}>application/json</code> and select the{" "}
                <strong>push</strong> event. Every push to the default branch then starts a deployment.
              </p>
              {[
                { label: "Payload URL", key: "url", value: webhook.url },
                { label: "Secret", key: "secret", value: webhook.secret },
              ].map((item) => (
                <div key={item.key} style={{ marginBottom: 18 }}>
                  <label className={ui.fieldLabel} htmlFor={`webhook-${item.key}`}>
                    {item.label}
                  </label>
                  <div className={ui.inlineForm}>
                    <input
                      id={`webhook-${item.key}`}
                      readOnly
                      value={item.value}
                      onFocus={(e) => e.currentTarget.select()}
                      className={`${ui.input} ${ui.inlineFormGrow} ${ui.mono}`}
                      style={{ fontSize: 13, background: "#fafafa" }}
                    />
                    <button
                      onClick={() => copyToClipboard(item.key, item.value)}
                      className={`${ui.btn} ${ui.btnSmall} ${ui.btnGhost}`}
                    >
                      {copiedField === item.key ? "Copied" : "Copy"}
                    </button>
                  </div>
                </div>
              ))}
              <p className={ui.note}>
                Keep the secret private — anyone with it can trigger deployments.
              </p>
            </>
          ) : (
            <p className={ui.empty}>Webhook details are unavailable.</p>
          )}
        </SidePanel>
      </main>
      </AppShell>
    </AuthGate>
  )
}
