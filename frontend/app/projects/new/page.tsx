"use client"

import { useEffect, useState, useCallback, useMemo, useRef } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import AppShell from "@/components/AppShell"
import AuthGate from "@/components/AuthGate"
import { apiFetch } from "@/lib/api"
import { safeExternalUrl } from "@/lib/url"
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
  const [repoCheckResult, setRepoCheckResult] = useState<{items?:{name:string;type:string;size?:number}[];readme?:string} | null>(null)
  const [repoCheckLoading, setRepoCheckLoading] = useState(false)
  const [repoChecked, setRepoChecked] = useState(false)
  const [manualContinue, setManualContinue] = useState(false)
  const [manualPreviewTab, setManualPreviewTab] = useState<"code" | "readme">("code")
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
  const hasSource = source === "zip" ? Boolean(zipFile) : hasGitSource
  const canDeploy =
    hasSource &&
    (source !== "manual" || manualContinue) &&
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
                  <span className={styles.tabIcon} aria-hidden="true">
                    {tab.key === "github" && (
                      <svg width="15" height="15" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12 .5C5.73.5.5 5.73.5 12c0 5.08 3.29 9.39 7.86 10.91.58.1.79-.25.79-.56v-2.17c-3.2.7-3.88-1.37-3.88-1.37-.53-1.34-1.29-1.7-1.29-1.7-1.05-.72.08-.7.08-.7 1.16.08 1.77 1.19 1.77 1.19 1.03 1.77 2.7 1.26 3.36.96.1-.75.4-1.26.73-1.55-2.55-.29-5.24-1.28-5.24-5.7 0-1.26.45-2.29 1.19-3.1-.12-.29-.52-1.46.11-3.05 0 0 .97-.31 3.18 1.18a11 11 0 0 1 5.79 0c2.2-1.49 3.17-1.18 3.17-1.18.63 1.59.23 2.76.12 3.05.74.81 1.18 1.84 1.18 3.1 0 4.43-2.69 5.41-5.25 5.7.41.35.78 1.05.78 2.12v3.14c0 .31.21.67.8.56A11.51 11.51 0 0 0 23.5 12C23.5 5.73 18.27.5 12 .5z" />
                      </svg>
                    )}
                    {tab.key === "manual" && (
                      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
                        <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
                      </svg>
                    )}
                    {tab.key === "zip" && (
                      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                        <polyline points="17 8 12 3 7 8" />
                        <line x1="12" y1="3" x2="12" y2="15" />
                      </svg>
                    )}
                  </span>
                  {tab.label}
                </button>
              ))}
            </div>

            {source === "github" ? (
              <div>
                {!loadingRepos && !reposError && repos.length > 0 && (
                  <div className={styles.searchWrap}>
                    <div className={styles.searchBox}>
                      <input
                        type="text"
                        className={styles.search}
                        value={repoQuery}
                        onChange={(e) => setRepoQuery(e.target.value)}
                        placeholder="Search repositories"
                        aria-label="Search repositories"
                      />
                      <span className={styles.searchIcon} aria-hidden="true">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          <circle cx="11" cy="11" r="7" />
                          <line x1="21" y1="21" x2="16.65" y2="16.65" />
                        </svg>
                      </span>
                    </div>
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
                        const repoLink = safeExternalUrl(r.html_url)
                        return (
                          <div
                            key={r.full_name}
                            role="option"
                            tabIndex={0}
                            aria-selected={active}
                            onClick={() => {
                              setSelectedRepo(r)
                              setPreviewRepo(r)
                              loadRepoContents(r)
                            }}
                            onKeyDown={(e) => {
                              if (e.key === "Enter" || e.key === " ") {
                                e.preventDefault()
                                setSelectedRepo(r)
                                setPreviewRepo(r)
                                loadRepoContents(r)
                              }
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
                            {repoLink && (
                              <a
                                href={repoLink}
                                target="_blank"
                                rel="noreferrer"
                                className={styles.repoOpen}
                                aria-label={`Open ${r.full_name} on GitHub`}
                                onClick={(e) => e.stopPropagation()}
                              >
                                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                                  <line x1="7" y1="17" x2="17" y2="7" />
                                  <polyline points="7 7 17 7 17 17" />
                                </svg>
                              </a>
                            )}
                          </div>
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
            ) : source === "manual" ? (
              <div className={styles.manualRepoRow} style={{ padding: "8px 0" }}>
                <label className={styles.manualRepoLabel} style={{ fontWeight: 600, fontSize: 14, color: "#171717", marginBottom: 8, display: "block" }}>Paste repository URL (private repo)</label>
                <div className={styles.inputWrap}>
                  <input
                    className={styles.input}
                    placeholder="https://github.com/username/repo-name.git"
                    value={manualRepoUrl}
                    onChange={(e) => setManualRepoUrl(e.target.value)}
                  />
                </div>
                <button
                  type="button"
                  className={`${styles.btn} ${styles.btnSecondary}`}
                  style={{ marginTop: 12, minWidth: 140, padding: "10px 16px", borderRadius: 8, border: "1px solid #ddd", background: "#fff", cursor: "pointer" }}
                  disabled={!manualRepoUrl.trim() || repoCheckLoading || repoChecked}
                  onClick={async () => {
                    setRepoCheckLoading(true)
                    setRepoCheckResult(null)
                    setManualPreviewTab("code")
                    try {
                      const url = new URL(manualRepoUrl.trim())
                      const path = url.pathname.replace(/\.git$/i, "").replace(/^\//, "")
                      const [owner, repo] = path.split("/")
                      if (!owner || !repo) throw new Error("Invalid URL")
                      const [contentsRes, readmeRes] = await Promise.all([
                        apiFetch(`/api/github/repos/${owner}/${repo}/contents`),
                        apiFetch(`/api/github/repos/${owner}/${repo}/readme`),
                      ])
                      const data = contentsRes.ok ? await contentsRes.json() : {}
                      const raw = data.contents || []
                      const items = Array.isArray(raw) ? raw.map((i:any)=>({name:String(i.name||""),type:String(i.type||""),size:Number(i.size||0)})) : []
                      let readme = ""
                      if (readmeRes.ok) {
                        const rd = await readmeRes.json()
                        readme = String(rd.readme || "")
                      }
                      setRepoCheckResult({ items, readme })
                      setRepoChecked(true)
                    } catch (err: any) {
                      setRepoCheckResult({ items: [], readme: "" })
                      setRepoChecked(true)
                      console.error("Repo check failed:", err)
                    } finally {
                      setRepoCheckLoading(false)
                    }
                  }}
                >
                  {repoCheckLoading ? "Checking..." : repoChecked ? "Checked" : "Check repo"}
                </button>
                {repoChecked && repoCheckResult && (
                  <div style={{ marginTop: 16, border: "1px solid #e9e9e9", borderRadius: 8, overflow: "hidden" }}>
                    <div style={{ display: "flex", gap: 4, borderBottom: "1px solid #e9e9e9", background: "#fafafa", padding: "0 8px" }}>
                      <button
                        type="button"
                        onClick={() => setManualPreviewTab("code")}
                        style={{
                          padding: "8px 16px", border: "none",
                          borderBottom: manualPreviewTab === "code" ? "2px solid #171717" : "2px solid transparent",
                          background: "transparent", cursor: "pointer", fontSize: "11pt",
                          fontWeight: manualPreviewTab === "code" ? 600 : 400,
                          color: manualPreviewTab === "code" ? "#171717" : "#8a8a8a",
                          fontFamily: "inherit",
                        }}
                      >Code</button>
                      <button
                        type="button"
                        onClick={() => setManualPreviewTab("readme")}
                        style={{
                          padding: "8px 16px", border: "none",
                          borderBottom: manualPreviewTab === "readme" ? "2px solid #171717" : "2px solid transparent",
                          background: "transparent", cursor: "pointer", fontSize: "11pt",
                          fontWeight: manualPreviewTab === "readme" ? 600 : 400,
                          color: manualPreviewTab === "readme" ? "#171717" : "#8a8a8a",
                          fontFamily: "inherit",
                        }}
                      >Readme</button>
                    </div>
                    <div style={{ padding: 12 }}>
                      {manualPreviewTab === "code" ? (
                        repoCheckResult.items && repoCheckResult.items.length > 0 ? (
                          <div style={{ border: "1px solid #e9e9e9", borderRadius: 8, overflow: "hidden" }}>
                            {repoCheckResult.items.map((item, idx) => (
                              <div key={item.name} style={{ display: "flex", alignItems: "center", gap: 8, padding: "8px 12px", borderBottom: idx < (repoCheckResult.items?.length || 0) - 1 ? "1px solid #f0f0f0" : "none", fontSize: "11pt" }}>
                                <span style={{ color: item.type === "dir" ? "#2563eb" : "#8a8a8a", fontSize: 14 }}>{item.type === "dir" ? "📁" : "📄"}</span>
                                <span style={{ flex: 1, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{item.name}</span>
                                {item.type === "file" && item.size ? (
                                  <span style={{ fontSize: "9pt", color: "#aaa" }}>{(item.size / 1024).toFixed(1)} KB</span>
                                ) : null}
                              </div>
                            ))}
                          </div>
                        ) : (
                          <p style={{ color: "#8a8a8a", fontSize: "11pt", margin: 0 }}>No files visible (repo might be private or empty).</p>
                        )
                      ) : repoCheckResult.readme ? (
                        <pre style={{
                          background: "#f8f8f8", border: "1px solid #e9e9e9", borderRadius: 8, padding: 16,
                          fontSize: "10pt", lineHeight: 1.6, overflow: "auto", maxHeight: "40vh",
                          whiteSpace: "pre-wrap", wordBreak: "break-word",
                          fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace", margin: 0,
                        }}>{repoCheckResult.readme}</pre>
                      ) : (
                        <p style={{ color: "#8a8a8a", fontSize: "11pt", margin: 0 }}>No README found in this repository.</p>
                      )}
                    </div>
                    {!manualContinue && (
                      <div style={{ padding: "0 12px 12px" }}>
                        <button
                          type="button"
                          style={{ padding: "10px 16px", borderRadius: 8, background: "#171717", color: "#fff", border: "none", cursor: "pointer", fontWeight: 600, fontSize: 13 }}
                          onClick={() => setManualContinue(true)}
                        >
                          Continue
                        </button>
                      </div>
                    )}
                  </div>
                )}
                {manualContinue && (
                  <p className={styles.manualRepoHint} style={{ marginTop: 12, color: "#1a7f37", fontWeight: 600 }}>
                    Continue — fill the subdomain on the right and click Deploy.
                  </p>
                )}
                <p className={styles.manualRepoHint} style={{ marginTop: 8 }}>Build will clone this private repository using your GitHub token.</p>
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
                    disabled={source === "manual" && !manualContinue}
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
                    {source === "zip" ? (
                      zipFile ? (
                        <strong>{zipFile.name}</strong>
                      ) : (
                        <span className={styles.summaryEmpty}>No ZIP chosen</span>
                      )
                    ) : selectedRepo ? (
                      <strong>{selectedRepo.full_name}</strong>
                    ) : manualRepoUrl.trim() ? (
                      <strong>{manualRepoUrl.trim()}</strong>
                    ) : (
                      <span className={styles.summaryEmpty}>No repository selected</span>
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
