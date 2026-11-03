"use client"

import { useEffect, useState, useCallback, useMemo, useRef } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import AppShell from "@/components/AppShell"
import AuthGate from "@/components/AuthGate"
import { apiFetch } from "@/lib/api"
import { useCurrentUser } from "@/lib/auth"
import { validateProjectName } from "@/lib/check-word"
import styles from "./new.module.css"
import ui from "../../ui.module.css"

interface Repo {
  name: string
  full_name: string
  html_url: string
  private: boolean
}

export default function NewProjectPage() {
  const router = useRouter()
  // Sesi diambil dari cache bersama; AuthGate yang menjaga halaman ini.
  const { user } = useCurrentUser()
  const [repos, setRepos] = useState<Repo[]>([])
  const [selectedRepo, setSelectedRepo] = useState<Repo | null>(null)
  const [subdomain, setSubdomain] = useState("")
  const [subdomainStatus, setSubdomainStatus] = useState<"idle" | "checking" | "available" | "taken" | "invalid">("idle")
  const [subdomainError, setSubdomainError] = useState("")
  const [loadingRepos, setLoadingRepos] = useState(true)
  const [reposError, setReposError] = useState("")
  const [repoPage, setRepoPage] = useState(1)
  const [hasMoreRepos, setHasMoreRepos] = useState(false)
  const [loadingMoreRepos, setLoadingMoreRepos] = useState(false)
  // Sumber deploy: repo GitHub atau unggahan ZIP.
  const [source, setSource] = useState<"github" | "manual" | "zip">("github")
  const [zipFile, setZipFile] = useState<File | null>(null)
  const [dragOver, setDragOver] = useState(false)
  const zipInputRef = useRef<HTMLInputElement>(null)
  const [submitting, setSubmitting] = useState(false)
  const [formError, setFormError] = useState("")
  const [repoQuery, setRepoQuery] = useState("")
  const [manualRepoUrl, setManualRepoUrl] = useState("")
  // Repo yang sudah punya project hidup: deploy ulang diblokir lewat modal info,
  // bukan error merah kecil di bawah form.
  const [duplicate, setDuplicate] = useState<{ id: string; subdomain: string } | null>(null)
  const [envVars, setEnvVars] = useState<{ key: string; value: string }[]>([])
  const [previewRepo, setPreviewRepo] = useState<Repo | null>(null)
  const [repoContents, setRepoContents] = useState<{ name: string; type: string; size: number; path: string }[]>([])
  const [loadingContents, setLoadingContents] = useState(false)
  const [previewTab, setPreviewTab] = useState<"code" | "readme">("code")
  const [readmeContent, setReadmeContent] = useState("")
  const [loadingReadme, setLoadingReadme] = useState(false)

  useEffect(() => {
    if (!duplicate) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setDuplicate(null)
    }
    document.addEventListener("keydown", onKey)
    document.body.style.overflow = "hidden"
    return () => {
      document.removeEventListener("keydown", onKey)
      document.body.style.overflow = ""
    }
  }, [duplicate])

  // Filter lokal: daftar repo bisa sampai 100 entri, jadi pencarian jauh lebih
  // efisien daripada menggulir.
  const visibleRepos = useMemo(() => {
    const q = repoQuery.trim().toLowerCase()
    if (!q) return repos
    return repos.filter((r) => r.name.toLowerCase().includes(q) || r.full_name.toLowerCase().includes(q))
  }, [repos, repoQuery])

  const hasGitSource = Boolean(selectedRepo || manualRepoUrl)
  const hasSource = source === "github" ? hasGitSource : Boolean(zipFile)
  const canDeploy =
    hasSource &&
    Boolean(subdomain) &&
    subdomainStatus !== "taken" &&
    subdomainStatus !== "invalid" &&
    subdomainStatus !== "checking"

  useEffect(() => {
    const load = async () => {
      const reposRes = await apiFetch("/api/github/repos?page=1")
      if (reposRes.ok) {
        const data = await reposRes.json()
        setRepos(data.repos || [])
        setHasMoreRepos(Boolean(data.has_more))
        setRepoPage(data.page || 1)
      } else {
        const txt = await reposRes.text()
        let msg = txt
        try {
          msg = JSON.parse(txt).error || txt
        } catch {}
        setReposError(msg || `Could not load repositories (${reposRes.status})`)
      }
      setLoadingRepos(false)
    }
    load()
  }, [router])

  const loadMoreRepos = async () => {
    if (loadingMoreRepos || !hasMoreRepos) return
    setLoadingMoreRepos(true)
    try {
      const nextPage = repoPage + 1
      const res = await apiFetch(`/api/github/repos?page=${nextPage}`)
      if (!res.ok) return
      const data = await res.json()
      const incoming: Repo[] = data.repos || []
      setRepos((prev) => {
        const seen = new Set(prev.map((r) => r.full_name))
        return [...prev, ...incoming.filter((r) => !seen.has(r.full_name))]
      })
      setRepoPage(data.page || nextPage)
      setHasMoreRepos(Boolean(data.has_more))
    } finally {
      setLoadingMoreRepos(false)
    }
  }

  // Check subdomain availability with debounce
  const checkSubdomain = useCallback(async (name: string) => {
    if (!name) {
      setSubdomainStatus("idle")
      setSubdomainError("")
      return
    }
    setSubdomainStatus("checking")
    setSubdomainError("")
    try {
      const res = await apiFetch(`/api/check-subdomain?name=${encodeURIComponent(name)}`)
      const data = await res.json()
      if (data.available) {
        setSubdomainStatus("available")
        setSubdomainError("")
      } else {
        setSubdomainStatus("taken")
        setSubdomainError(data.error || "This name is already taken. Please choose another.")
      }
    } catch {
      setSubdomainStatus("idle")
      setSubdomainError("")
    }
  }, [])

  useEffect(() => {
    if (!subdomain) return
    const result = validateProjectName(subdomain)
    if (!result.valid) return
    const timer = setTimeout(() => {
      checkSubdomain(subdomain)
    }, 400)
    return () => clearTimeout(timer)
  }, [subdomain, checkSubdomain])

  const loadRepoContents = async (repo: Repo) => {
    const parts = repo.full_name.split("/")
    if (parts.length < 2) return
    setLoadingContents(true)
    setLoadingReadme(true)
    setRepoContents([])
    setReadmeContent("")
    setPreviewTab("code")
    try {
      const [contentsRes, readmeRes] = await Promise.all([
        apiFetch(`/api/github/repos/${parts[0]}/${parts[1]}/contents`),
        apiFetch(`/api/github/repos/${parts[0]}/${parts[1]}/readme`),
      ])
      if (contentsRes.ok) {
        const data = await contentsRes.json()
        setRepoContents(data.contents || [])
      }
      if (readmeRes.ok) {
        const data = await readmeRes.json()
        setReadmeContent(data.readme || "")
      }
    } catch {
      // ignore
    } finally {
      setLoadingContents(false)
      setLoadingReadme(false)
    }
  }

  const handleSubdomainChange = (value: string) => {
    const cleaned = value.toLowerCase().replace(/[^a-z0-9-]/g, "")
    setSubdomain(cleaned)
    if (!cleaned) {
      setSubdomainStatus("idle")
      setSubdomainError("")
      return
    }
    const result = validateProjectName(cleaned)
    if (!result.valid) {
      setSubdomainStatus("invalid")
      setSubdomainError(result.error)
    } else {
      setSubdomainStatus("checking")
      setSubdomainError("")
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (source === "github" && !selectedRepo && !manualRepoUrl.trim()) {
      setFormError("Please select a repository or paste a repository URL.")
      return
    }
    if (source === "zip" && !zipFile) {
      setFormError("Please choose a ZIP file first.")
      return
    }
    if (!subdomain) {
      setFormError("Please enter a project name.")
      return
    }
    const validation = validateProjectName(subdomain)
    if (!validation.valid) {
      setFormError(validation.error)
      return
    }
    if (subdomainStatus === "taken" || subdomainStatus === "invalid") {
      setFormError("Please fix the project name before deploying.")
      return
    }
    const envObj: Record<string, string> = {}
    for (const { key, value } of envVars) {
      if (key.trim()) envObj[key.trim()] = value
    }
    setFormError("")

    // Repo yang sudah pernah di-deploy: beri tahu lewat modal agar pengguna
    // menghapus project lama dulu, bukan membuat salinan yang membingungkan.
    if (source === "github") {
      const norm = (u: string) => u.trim().toLowerCase().replace(/\.git$/, "").replace(/\/+$/, "")
      try {
        const listRes = await apiFetch("/api/projects")
        if (listRes.ok) {
          const listData = await listRes.json()
          const target = norm(manualRepoUrl.trim() || selectedRepo!.html_url)
          const hit = (listData.projects || []).find(
            (p: { id: string; repo_url?: string; subdomain?: string }) => p.repo_url && norm(p.repo_url) === target
          )
          if (hit) {
            setDuplicate({ id: hit.id, subdomain: hit.subdomain })
            return
          }
        }
      } catch {
        // Pemeriksaan gagal jangan menghalangi deploy — biarkan backend yang menilai.
      }
    }

    setSubmitting(true)
    try {
      let res: Response
      if (source === "zip") {
        const form = new FormData()
        form.append("subdomain", subdomain)
        form.append("file", zipFile as File)
        res = await apiFetch("/api/projects/upload", { method: "POST", body: form })
      } else {
        res = await apiFetch("/api/projects", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ repo_url: manualRepoUrl.trim() || selectedRepo!.html_url, subdomain, env_vars: envObj }),
        })
      }
      if (!res.ok) {
        const txt = await res.text()
        let msg = txt
        try {
          msg = JSON.parse(txt).error || txt
        } catch {}
        setFormError(msg || `Deploy failed (${res.status})`)
        return
      }
      const data = await res.json()
      router.push(`/projects/${data.project.id}`)
    } catch (err: any) {
      setFormError(String(err?.message || err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthGate>
      <AppShell user={user}>
      <main className={`${ui.page} ${ui.pageWide}`}>
        <div className={ui.topbar}>
          <Link href="/dashboard" className={ui.back}>
            ← Dashboard
          </Link>
          <span className={ui.eyebrow}>New project</span>
        </div>

        <header className={ui.head}>
          <div className={ui.headMain}>
            <h1 className={ui.title}>Ship something fun!</h1>
            <p className={ui.lede}>
              Pick a repository or drop a static archive, name it, and it goes live on its own subdomain under
              only4.fun.
            </p>
          </div>
        </header>

        <div className={styles.workspace}>
          {/* Langkah 1: sumber */}
          <section className={styles.col}>
            <div className={styles.sectionHead}>
              <h2 className={styles.sectionTitle}>Source</h2>
            </div>
            <p className={styles.sectionHint}>Where the files come from.</p>

            <div className={styles.tabs} role="tablist" aria-label="Deploy source">
              {[
                { key: "github" as const, label: "GitHub repository" },
                { key: "manual" as const, label: "Paste URL" },
                { key: "zip" as const, label: "Upload ZIP" },
              ].map((tab) => (
                <button
                  key={tab.key}
                  type="button"
                  role="tab"
                  aria-selected={source === tab.key}
                  className={`${styles.tab} ${source === tab.key ? styles.tabActive : ""}`}
                  onClick={() => {
                    setSource(tab.key)
                    setFormError("")
                  }}
                >
                  {tab.label}
                </button>
              ))}
            </div>

            {source === "github" ? (
              <div>
                {!loadingRepos && !reposError && repos.length > 0 && (
                  <div className={styles.searchWrap}>
                    <input
                      type="text"
                      className={styles.search}
                      value={repoQuery}
                      onChange={(e) => setRepoQuery(e.target.value)}
                      placeholder="Search repositories"
                      aria-label="Search repositories"
                    />
                    <div className={styles.searchMeta}>
                      <span>
                        {visibleRepos.length} of {repos.length} repositories
                      </span>
                      <span>GitHub</span>
                    </div>
                  </div>
                )}

                {loadingRepos ? (
                  <p className={styles.empty}>Loading repositories...</p>
                ) : reposError ? (
                  <p className={styles.errorText}>{reposError}</p>
                ) : repos.length === 0 ? (
                  <p className={styles.empty}>No repositories found.</p>
                ) : visibleRepos.length === 0 ? (
                  <p className={styles.empty}>No repositories match &quot;{repoQuery}&quot;.</p>
                ) : (
                  <>
                    <div className={styles.repoList} role="listbox" aria-label="Repositories">
                      {visibleRepos.map((r) => {
                        const active = selectedRepo?.full_name === r.full_name
                        return (
                          <button
                            key={r.full_name}
                            type="button"
                            role="option"
                            aria-selected={active}
                            onClick={() => {
                              setSelectedRepo(r)
                              setPreviewRepo(r)
                              loadRepoContents(r)
                            }}
                            className={`${styles.repoRow} ${active ? styles.repoRowActive : ""}`}
                          >
                            <div className={styles.repoText}>
                              <div className={styles.repoName}>{r.name}</div>
                              <div className={styles.repoMeta}>
                                {r.full_name} · {r.private ? "private" : "public"}
                              </div>
                            </div>
                            {active && <span className={styles.check}>✓</span>}
                          </button>
                        )
                      })}
                    </div>
                    {hasMoreRepos && !repoQuery.trim() && (
                      <button type="button" className={styles.more} onClick={loadMoreRepos} disabled={loadingMoreRepos}>
                        {loadingMoreRepos ? "Loading..." : "Load more repositories"}
                      </button>
                    )}
                  </>
                )}
              </div>
            ) : (
              <>
                <div
                  className={`${styles.dropzone} ${dragOver ? styles.dropzoneActive : ""}`}
                  role="button"
                  tabIndex={0}
                  aria-label="Choose a ZIP file"
                  onDragOver={(e) => {
                    e.preventDefault()
                    setDragOver(true)
                  }}
                  onDragLeave={() => setDragOver(false)}
                  onDrop={(e) => {
                    e.preventDefault()
                    setDragOver(false)
                    const dropped = e.dataTransfer.files?.[0]
                    if (dropped && dropped.name.toLowerCase().endsWith(".zip")) setZipFile(dropped)
                  }}
                  onClick={() => zipInputRef.current?.click()}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      e.preventDefault()
                      zipInputRef.current?.click()
                    }
                  }}
                >
                  <div className={styles.dropTitle}>
                    {zipFile ? zipFile.name : "Drop a .zip here, or click to choose"}
                  </div>
                  <div className={styles.dropHint}>
                    {zipFile ? `${(zipFile.size / (1024 * 1024)).toFixed(1)} MB` : "Maximum 100 MB"}
                  </div>
                  <input
                    ref={zipInputRef}
                    type="file"
                    accept=".zip,application/zip"
                    onChange={(e) => setZipFile(e.target.files?.[0] || null)}
                    style={{ display: "none" }}
                  />
                </div>
                <p className={styles.dropNote}>
                  The archive must contain <code className={styles.code}>index.html</code> at its root, or inside a
                  single top-level folder. Files are served as-is, no build step runs.
                </p>
              </>
            )}
          </section>

          {/* Langkah 2 & 3: nama dan review */}
          <section className={`${styles.col} ${styles.formCol}`}>
            <div className={styles.section}>
              <div className={styles.sectionHead}>
                <h2 className={styles.sectionTitle}>Project name</h2>
              </div>
              <p className={styles.sectionHint}>This becomes your public URL.</p>

              <div className={styles.field}>
                <label className={styles.fieldLabel} htmlFor="project-name">
                  Subdomain
                </label>
                <div className={styles.inputWrap}>
                  <input
                    id="project-name"
                    className={styles.input}
                    value={subdomain}
                    onChange={(e) => handleSubdomainChange(e.target.value)}
                    placeholder="myproject"
                    autoComplete="off"
                    spellCheck={false}
                    aria-describedby="project-name-status"
                  />
                  <span className={styles.suffix}>.only4.fun</span>
                </div>

                {subdomain ? (
                  <div id="project-name-status" className={styles.status}>
                    {subdomainStatus === "checking" && <span>Checking availability...</span>}
                    {subdomainStatus === "available" && (
                      <span className={styles.statusOk}>
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                          <polyline points="20 6 9 17 4 12" />
                        </svg>{" "}
                        Available
                      </span>
                    )}
                    {(subdomainStatus === "taken" || subdomainStatus === "invalid") && (
                      <span className={styles.statusErr}>
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                          <line x1="18" y1="6" x2="6" y2="18" />
                          <line x1="6" y1="6" x2="18" y2="18" />
                        </svg>{" "}
                        {subdomainError}
                      </span>
                    )}
                  </div>
                ) : (
                  <p className={styles.hint}>Lowercase letters, numbers and dashes only. At least 3 characters.</p>
                )}
              </div>
            </div>

            <div className={styles.section}>
              <div className={styles.sectionHead}>
                <h2 className={styles.sectionTitle}>Environment variables</h2>
              </div>
              <p className={styles.sectionHint}>Optional. Add key/value pairs needed at build time (e.g. API keys). Leave empty if not needed.</p>
              {envVars.map((ev, idx) => (
                <div key={idx} style={{ display: "flex", gap: 8, marginBottom: 8, alignItems: "center" }}>
                  <input
                    className={styles.input}
                    placeholder="KEY"
                    value={ev.key}
                    onChange={(e) => {
                      const next = [...envVars]
                      next[idx] = { ...next[idx], key: e.target.value }
                      setEnvVars(next)
                    }}
                    style={{ flex: 1 }}
                  />
                  <input
                    className={styles.input}
                    placeholder="value"
                    type="password"
                    value={ev.value}
                    onChange={(e) => {
                      const next = [...envVars]
                      next[idx] = { ...next[idx], value: e.target.value }
                      setEnvVars(next)
                    }}
                    style={{ flex: 2 }}
                  />
                  <button
                    type="button"
                    onClick={() => setEnvVars(envVars.filter((_, i) => i !== idx))}
                    style={{ background: "none", border: "none", cursor: "pointer", color: "#c5221f", fontSize: "16px", padding: "0 4px" }}
                    aria-label="Remove"
                  >
                    ×
                  </button>
                </div>
              ))}
              <button
                type="button"
                onClick={() => setEnvVars([...envVars, { key: "", value: "" }])}
                className={styles.more}
                style={{ marginTop: 4 }}
              >
                + Add variable
              </button>
            </div>

            <div className={styles.section}>
              <div className={styles.sectionHead}>
                <h2 className={styles.sectionTitle}>Review</h2>
              </div>
              <p className={styles.sectionHint}>One last look before it goes live.</p>

              <div className={styles.summary}>
                <div className={styles.summaryRow}>
                  <span className={styles.summaryLabel}>Source</span>
                  <span className={styles.summaryValue}>
                    {source === "github" ? (
                      selectedRepo ? (
                        <strong>{selectedRepo.full_name}</strong>
                      ) : (
                        <span className={styles.summaryEmpty}>No repository selected</span>
                      )
                    ) : zipFile ? (
                      <strong>{zipFile.name}</strong>
                    ) : (
                      <span className={styles.summaryEmpty}>No ZIP chosen</span>
                    )}
                  </span>
                </div>
                <div className={styles.summaryRow}>
                  <span className={styles.summaryLabel}>Live URL</span>
                  <span className={styles.summaryValue}>
                    {subdomain ? (
                      <strong>{subdomain}.only4.fun</strong>
                    ) : (
                      <span className={styles.summaryEmpty}>Choose a name</span>
                    )}
                  </span>
                </div>
                <div className={styles.summaryRow}>
                  <span className={styles.summaryLabel}>Environment</span>
                  <span className={styles.summaryValue}>
                    {envVars.filter((e) => e.key.trim()).length > 0
                      ? `${envVars.filter((e) => e.key.trim()).length} variable(s) set`
                      : "None (add after deploy if needed)"}
                  </span>
                </div>
              </div>

              {formError && (
                <p className={styles.formError} role="alert">
                  <span className={styles.dot} aria-hidden="true" />
                  {formError}
                </p>
              )}

              <form onSubmit={handleSubmit}>
                <div className={styles.actions}>
                  <Link href="/dashboard" className={styles.cancel}>
                    Cancel
                  </Link>
                  <button type="submit" className={styles.deploy} disabled={!canDeploy || submitting}>
                    {submitting ? "Deploying..." : "Deploy project"}
                  </button>
                </div>
              </form>

              <p className={styles.footNote}>Environment variables can be added after deployment from the project page.</p>
            </div>
          </section>
        </div>

        {previewRepo && (
          <div
            role="dialog"
            aria-modal="true"
            aria-label="Repository preview"
            onClick={(e) => {
              if (e.target === e.currentTarget) setPreviewRepo(null)
            }}
            style={{
              position: "fixed",
              inset: 0,
              zIndex: 9999,
              background: "rgba(0,0,0,0.45)",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              padding: 24,
            }}
          >
            <div
              style={{
                background: "#fff",
                borderRadius: 16,
                padding: 24,
                maxWidth: 560,
                width: "100%",
                maxHeight: "80vh",
                overflow: "auto",
                boxShadow: "0 20px 60px rgba(0,0,0,0.3)",
                fontFamily: "Inter, system-ui, sans-serif",
              }}
            >
              <h3 style={{ fontSize: "16pt", fontWeight: 600, margin: "0 0 4px", color: "#171717" }}>
                Repository Preview
              </h3>
              <p style={{ fontSize: "10pt", color: "#8a8a8a", margin: "0 0 16px" }}>
                {previewRepo.full_name} · {previewRepo.private ? "private" : "public"}
              </p>

              <div style={{ display: "flex", gap: 4, marginBottom: 16, borderBottom: "1px solid #e9e9e9" }}>
                <button
                  type="button"
                  onClick={() => setPreviewTab("code")}
                  style={{
                    padding: "8px 16px",
                    border: "none",
                    borderBottom: previewTab === "code" ? "2px solid #171717" : "2px solid transparent",
                    background: "transparent",
                    cursor: "pointer",
                    fontSize: "11pt",
                    fontWeight: previewTab === "code" ? 600 : 400,
                    color: previewTab === "code" ? "#171717" : "#8a8a8a",
                    fontFamily: "inherit",
                  }}
                >
                  Code
                </button>
                <button
                  type="button"
                  onClick={() => setPreviewTab("readme")}
                  style={{
                    padding: "8px 16px",
                    border: "none",
                    borderBottom: previewTab === "readme" ? "2px solid #171717" : "2px solid transparent",
                    background: "transparent",
                    cursor: "pointer",
                    fontSize: "11pt",
                    fontWeight: previewTab === "readme" ? 600 : 400,
                    color: previewTab === "readme" ? "#171717" : "#8a8a8a",
                    fontFamily: "inherit",
                  }}
                >
                  Readme
                </button>
              </div>

              {previewTab === "code" ? (
                loadingContents ? (
                  <p style={{ color: "#8a8a8a", fontSize: "11pt" }}>Loading contents...</p>
                ) : repoContents.length === 0 ? (
                  <p style={{ color: "#8a8a8a", fontSize: "11pt" }}>No files found or empty repository.</p>
                ) : (
                  <div style={{ border: "1px solid #e9e9e9", borderRadius: 8, overflow: "hidden", marginBottom: 16 }}>
                    {repoContents.map((item, idx) => (
                      <div
                        key={item.path}
                        style={{
                          display: "flex",
                          alignItems: "center",
                          gap: 8,
                          padding: "8px 12px",
                          borderBottom: idx < repoContents.length - 1 ? "1px solid #f0f0f0" : "none",
                          fontSize: "11pt",
                        }}
                      >
                        <span style={{ color: item.type === "dir" ? "#2563eb" : "#8a8a8a", fontSize: "14px" }}>
                          {item.type === "dir" ? "📁" : "📄"}
                        </span>
                        <span style={{ flex: 1, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
                          {item.name}
                        </span>
                        {item.type === "file" && (
                          <span style={{ fontSize: "9pt", color: "#aaa" }}>
                            {(item.size / 1024).toFixed(1)} KB
                          </span>
                        )}
                      </div>
                    ))}
                  </div>
                )
              ) : (
                loadingReadme ? (
                  <p style={{ color: "#8a8a8a", fontSize: "11pt" }}>Loading readme...</p>
                ) : readmeContent ? (
                  <pre style={{
                    background: "#f8f8f8",
                    border: "1px solid #e9e9e9",
                    borderRadius: 8,
                    padding: 16,
                    fontSize: "10pt",
                    lineHeight: 1.6,
                    overflow: "auto",
                    maxHeight: "40vh",
                    whiteSpace: "pre-wrap",
                    wordBreak: "break-word",
                    fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
                    marginBottom: 16,
                  }}>
                    {readmeContent}
                  </pre>
                ) : (
                  <p style={{ color: "#8a8a8a", fontSize: "11pt" }}>No README found in this repository.</p>
                )
              )}

              <p style={{ fontSize: "10pt", color: "#8a8a8a", margin: "0 0 16px" }}>
                Review the files above. If everything looks correct, proceed to deploy.
              </p>

              <div style={{ display: "flex", gap: 10, justifyContent: "flex-end" }}>
                <button
                  type="button"
                  onClick={() => {
                    setPreviewRepo(null)
                    setSelectedRepo(null)
                  }}
                  style={{
                    padding: "8px 14px",
                    borderRadius: 8,
                    border: "1px solid #ddd",
                    background: "#fff",
                    fontSize: "10pt",
                    cursor: "pointer",
                    fontFamily: "inherit",
                  }}
                >
                  Cancel
                </button>
                <button
                  type="button"
                  onClick={() => setPreviewRepo(null)}
                  style={{
                    padding: "8px 14px",
                    borderRadius: 8,
                    border: "none",
                    background: "#171717",
                    color: "#fff",
                    fontSize: "10pt",
                    fontWeight: 600,
                    cursor: "pointer",
                    fontFamily: "inherit",
                  }}
                >
                  Continue to deploy
                </button>
              </div>
            </div>
          </div>
        )}

        {duplicate && (
          <div
            role="dialog"
            aria-modal="true"
            aria-label="Repository already deployed"
            onClick={(e) => {
              if (e.target === e.currentTarget) setDuplicate(null)
            }}
            style={{
              position: "fixed",
              inset: 0,
              zIndex: 9999,
              background: "rgba(0,0,0,0.45)",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              padding: 24,
            }}
          >
            <div
              style={{
                background: "#fff",
                borderRadius: 16,
                padding: 24,
                maxWidth: 420,
                width: "100%",
                boxShadow: "0 20px 60px rgba(0,0,0,0.3)",
                fontFamily: "Inter, system-ui, sans-serif",
              }}
            >
              <h3 style={{ fontSize: "16pt", fontWeight: 600, margin: "0 0 8px", color: "#171717" }}>
                Repository already deployed
              </h3>
              <p style={{ fontSize: "11pt", color: "#555", margin: "0 0 20px", lineHeight: 1.45 }}>
                This repository is already deployed as{" "}
                <strong style={{ color: "#171717" }}>{duplicate.subdomain}.only4.fun</strong>. Delete that project
                first, then deploy it again from this page.
              </p>
              <div style={{ display: "flex", gap: 10, justifyContent: "flex-end" }}>
                <button
                  type="button"
                  onClick={() => setDuplicate(null)}
                  style={{
                    padding: "8px 14px",
                    borderRadius: 8,
                    border: "1px solid #ddd",
                    background: "#fff",
                    fontSize: "10pt",
                    cursor: "pointer",
                    fontFamily: "inherit",
                  }}
                >
                  Close
                </button>
                <Link
                  href={`/projects/${duplicate.id}`}
                  style={{
                    padding: "8px 14px",
                    borderRadius: 8,
                    border: "none",
                    background: "#171717",
                    color: "#fff",
                    fontSize: "10pt",
                    fontWeight: 600,
                    textDecoration: "none",
                    fontFamily: "inherit",
                  }}
                >
                  Open project
                </Link>
              </div>
            </div>
          </div>
        )}
      </main>
      </AppShell>
    </AuthGate>
  )
}
