"use client"

import { useEffect, useState } from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import AppShell from "@/components/AppShell"
import AuthGate from "@/components/AuthGate"
import ContributionGraph, { ContributionDay } from "@/components/ContributionGraph"
import { apiFetch } from "@/lib/api"
import { optimizeReadmeHtml } from "@/lib/readme"
import { clearCurrentUser, useCurrentUser } from "@/lib/auth"
import ui from "../ui.module.css"
import styles from "./profile.module.css"

interface GitHubProfile {
  login: string
  name: string
  bio: string
  company: string
  location: string
  blog: string
  twitter: string
  public_repos: number
  followers: number
  following: number
  created_at: string
  readme_html: string
  contributions: ContributionDay[]
  total_contributions: number
}

function joined(iso: string) {
  if (!iso) return ""
  const parsed = new Date(iso)
  if (Number.isNaN(parsed.getTime())) return ""
  return parsed.toLocaleDateString("en-GB", { month: "long", year: "numeric" })
}

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className={ui.kvRow}>
      <span className={ui.kvLabel}>{label}</span>
      <span className={ui.kvValue}>{children}</span>
    </div>
  )
}

export default function ProfilePage() {
  const router = useRouter()
  // Sesi diambil dari cache bersama; AuthGate yang menjaga halaman ini.
  const { user, loaded } = useCurrentUser()
  const [loggingOut, setLoggingOut] = useState(false)
  const [error, setError] = useState("")
  const [profile, setProfile] = useState<GitHubProfile | null>(null)
  const [profileLoading, setProfileLoading] = useState(true)
  const [profileError, setProfileError] = useState("")

  // Data tambahan dari GitHub (bio, README, kalender kontribusi) dimuat
  // terpisah supaya halaman tetap tampil walau GitHub tidak merespons.
  useEffect(() => {
    if (!loaded || !user) return
    let active = true
    setProfileLoading(true)
    apiFetch("/api/github/profile")
      .then(async (res) => {
        if (!res.ok) {
          const txt = await res.text()
          let msg = txt
          try {
            msg = JSON.parse(txt).error || txt
          } catch {}
          throw new Error(msg || `Could not load your GitHub profile (${res.status})`)
        }
        return res.json()
      })
      .then((data) => {
        if (active) {
          setProfile(data.profile || null)
          setProfileError("")
        }
      })
      .catch((e: any) => {
        if (active) setProfileError(String(e?.message || e))
      })
      .finally(() => {
        if (active) setProfileLoading(false)
      })
    return () => {
      active = false
    }
  }, [loaded, user])

  const handleLogout = async () => {
    setLoggingOut(true)
    setError("")
    try {
      await apiFetch("/api/auth/logout", { method: "POST" })
    } catch {
      // tetap arahkan ke beranda walau backend tidak merespons
    }
    clearCurrentUser()
    router.push("/")
  }

  return (
    <AuthGate>
      <AppShell user={user}>
      <main className={`${ui.page} ${ui.pageWide}`}>
        <div className={ui.topbar}>
          <Link href="/dashboard" className={ui.back}>
            ← Dashboard
          </Link>
          <span className={ui.eyebrow}>Profile</span>
        </div>

        <header className={ui.head}>
          <div className={ui.headMain}>
            <h1 className={ui.title}>Profile.</h1>
            <p className={ui.lede}>Your account, session, and the links you may need.</p>
          </div>
        </header>

        {error && (
          <p className={ui.errorText} role="alert">
            <span className={ui.dot} aria-hidden="true" />
            {error}
          </p>
        )}

        <div className={ui.reading}>
          {!loaded || !user ? (
            <p className={ui.empty}>Loading...</p>
          ) : (
            <>
              <section className={ui.section}>
                <div className={ui.sectionHead}>
                  <div className={ui.sectionHeadMain}>
                    <h2 className={ui.sectionTitle}>Account</h2>
                  </div>
                  <span className={ui.eyebrow}>GitHub</span>
                </div>

                <div style={{ display: "flex", alignItems: "center", gap: 16, marginTop: 16, flexWrap: "wrap" }}>
                  {user.avatar_url ? (
                    <img
                      src={user.avatar_url}
                      alt={user.username}
                      width={56}
                      height={56}
                      // Avatar utama ada di atas layar, jadi dimuat lebih dulu.
                      loading="eager"
                      decoding="async"
                      referrerPolicy="no-referrer"
                      style={{ width: 56, height: 56, borderRadius: "50%", border: "1px solid #e4e4e4" }}
                    />
                  ) : (
                    <div
                      style={{ width: 56, height: 56, borderRadius: "50%", background: "#f0f0f0" }}
                      aria-hidden="true"
                    />
                  )}
                  <div style={{ minWidth: 0 }}>
                    <div style={{ fontSize: 16, fontWeight: 650, marginBottom: 2, overflowWrap: "anywhere" }}>
                      {user.username}
                    </div>
                    <div style={{ fontSize: 13, color: "#5c5c5c" }}>Signed in with GitHub</div>
                  </div>
                </div>

                <div className={ui.kv} style={{ marginTop: 14 }}>
                  <Row label="Username">{user.username}</Row>
                  <Row label="Email">{user.email || "—"}</Row>
                  <Row label="GitHub ID">{user.github_id}</Row>
                </div>

                <a
                  href="https://github.com/settings/profile"
                  target="_blank"
                  rel="noreferrer"
                  className={ui.externalLink}
                >
                  Edit your profile on GitHub
                  <svg width={12} height={12} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                    <path d="M7 17 17 7" />
                    <path d="M9 7h8v8" />
                  </svg>
                </a>
              </section>

              <section className={ui.section}>
                <div className={ui.sectionHead}>
                  <div className={ui.sectionHeadMain}>
                    <h2 className={ui.sectionTitle}>About</h2>
                  </div>
                  {profile && (
                    <a
                      href={`https://github.com/${profile.login}`}
                      target="_blank"
                      rel="noreferrer"
                      className={ui.eyebrow}
                    >
                      github.com/{profile.login}
                    </a>
                  )}
                </div>

                {profileLoading ? (
                  <p className={ui.empty} style={{ marginTop: 14 }}>
                    Loading GitHub profile...
                  </p>
                ) : profileError ? (
                  <p className={ui.errorText} role="alert">
                    <span className={ui.dot} aria-hidden="true" />
                    {profileError}
                  </p>
                ) : profile ? (
                  <>
                    {profile.bio ? (
                      <p className={styles.bio}>{profile.bio}</p>
                    ) : (
                      <p className={styles.muted}>No bio on GitHub yet.</p>
                    )}

                    <div className={styles.meta}>
                      {profile.company && (
                        <div className={styles.metaItem}>
                          <span className={styles.metaLabel}>Company</span>
                          <span className={styles.metaValue}>{profile.company}</span>
                        </div>
                      )}
                      {profile.location && (
                        <div className={styles.metaItem}>
                          <span className={styles.metaLabel}>Location</span>
                          <span className={styles.metaValue}>{profile.location}</span>
                        </div>
                      )}
                      {profile.blog && (
                        <div className={styles.metaItem}>
                          <span className={styles.metaLabel}>Website</span>
                          <span className={styles.metaValue}>
                            <a
                              href={profile.blog.startsWith("http") ? profile.blog : `https://${profile.blog}`}
                              target="_blank"
                              rel="noreferrer"
                            >
                              {profile.blog.replace(/^https?:\/\//, "")}
                            </a>
                          </span>
                        </div>
                      )}
                      {profile.twitter && (
                        <div className={styles.metaItem}>
                          <span className={styles.metaLabel}>X / Twitter</span>
                          <span className={styles.metaValue}>@{profile.twitter}</span>
                        </div>
                      )}
                      {joined(profile.created_at) && (
                        <div className={styles.metaItem}>
                          <span className={styles.metaLabel}>Joined</span>
                          <span className={styles.metaValue}>{joined(profile.created_at)}</span>
                        </div>
                      )}
                      <div className={styles.metaItem}>
                        <span className={styles.metaLabel}>Public repos</span>
                        <span className={styles.metaValue}>{profile.public_repos.toLocaleString("en-GB")}</span>
                      </div>
                      <div className={styles.metaItem}>
                        <span className={styles.metaLabel}>Followers</span>
                        <span className={styles.metaValue}>{profile.followers.toLocaleString("en-GB")}</span>
                      </div>
                      <div className={styles.metaItem}>
                        <span className={styles.metaLabel}>Following</span>
                        <span className={styles.metaValue}>{profile.following.toLocaleString("en-GB")}</span>
                      </div>
                    </div>
                  </>
                ) : null}
              </section>

              <section className={ui.section}>
                <div className={ui.sectionHead}>
                  <div className={ui.sectionHeadMain}>
                    <h2 className={ui.sectionTitle}>Contributions</h2>
                  </div>
                </div>
                {profileLoading ? (
                  <p className={ui.empty} style={{ marginTop: 14 }}>
                    Loading activity...
                  </p>
                ) : profile ? (
                  <ContributionGraph days={profile.contributions || []} total={profile.total_contributions || 0} />
                ) : (
                  <p className={ui.empty} style={{ marginTop: 14 }}>
                    Contribution activity is unavailable right now.
                  </p>
                )}
              </section>

              <section className={ui.section}>
                <div className={ui.sectionHead}>
                  <div className={ui.sectionHeadMain}>
                    <h2 className={ui.sectionTitle}>README</h2>
                  </div>
                  {profile && <span className={ui.eyebrow}>From GitHub</span>}
                </div>
                {profileLoading ? (
                  <p className={ui.empty} style={{ marginTop: 14 }}>
                    Loading README...
                  </p>
                ) : profile?.readme_html ? (
                  <div className={styles.readme} dangerouslySetInnerHTML={{ __html: optimizeReadmeHtml(profile.readme_html) }} />
                ) : (
                  <p className={ui.empty} style={{ marginTop: 14 }}>
                    No profile README found. Create a repository named <strong>{user.username}</strong> with a
                    README.md to show one here.
                  </p>
                )}
              </section>

              <section className={ui.section}>
                <div className={ui.sectionHead}>
                  <div className={ui.sectionHeadMain}>
                    <h2 className={ui.sectionTitle}>Session</h2>
                  </div>
                </div>
                <p className={ui.sectionHint}>
                  Signing out deletes the session on the server and clears the cookie in this browser.
                </p>

                <div className={ui.sectionBody} style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
                  <Link href="/help" className={`${ui.btn} ${ui.btnGhost}`}>
                    Help
                  </Link>
                  <Link href="/support" className={`${ui.btn} ${ui.btnGhost}`}>
                    Support
                  </Link>
                  <button
                    type="button"
                    onClick={handleLogout}
                    disabled={loggingOut}
                    className={`${ui.btn} ${ui.btnDanger}`}
                  >
                    {loggingOut ? "Logging out..." : "Logout"}
                  </button>
                </div>
              </section>
            </>
          )}
        </div>
      </main>
      </AppShell>
    </AuthGate>
  )
}
