"use client"

import Link from "next/link"
import AppShell from "@/components/AppShell"
import { useCurrentUser } from "@/lib/auth"
import ui from "../ui.module.css"
import styles from "../content.module.css"

// Ganti dengan URL sebenarnya saat sudah tersedia.
const GITHUB_URL = "https://github.com/only4fun/only4.fun"
const SUPPORT_URL = "https://example.com/only4.fun/support"
const LICENSE_URL = "https://opensource.org/license/mit"

const H2: React.CSSProperties = { fontSize: "16px", fontWeight: 650, margin: "34px 0 12px", paddingTop: 24, borderTop: "1px solid #e4e4e4", letterSpacing: "-0.1px" }
const H3: React.CSSProperties = { fontSize: "14px", fontWeight: 650, margin: "24px 0 6px" }
const P: React.CSSProperties = { fontSize: "13.5px", lineHeight: 1.75, color: "#4f4f4f", margin: "0 0 12px" }
const UL: React.CSSProperties = { margin: "0 0 14px", paddingLeft: 22 }
const LI: React.CSSProperties = { fontSize: "13.5px", lineHeight: 1.75, color: "#4f4f4f", marginBottom: 6 }
const CODE: React.CSSProperties = {
  background: "#f4f4f4",
  padding: "1px 5px",
  borderRadius: 4,
  fontSize: "12.5px",
  fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
}

export default function SupportPage() {
  // Support bisa dibaca tanpa login: kalau sesi tidak ada, cukup jadi tamu.
  // Cache dari halaman sebelumnya dipakai agar shell tidak berkedip.
  const { user, loaded } = useCurrentUser()

  return (
    <AppShell user={user} loading={!loaded} publicShell>
      <main className={`${ui.page} ${ui.pageWide}`}>
        <div className={ui.topbar}>
          <Link href="/dashboard" className={ui.back}>
            ← Dashboard
          </Link>
          <span className={ui.eyebrow}>Support</span>
        </div>
        <header className={ui.head}>
          <div className={ui.headMain}>
            <h1 className={ui.title}>Support</h1>
            <p className={ui.lede}>
              only4.fun is a small project. If something is broken, unclear, or you want to lend a hand, here is where
              to go.
            </p>
          </div>
        </header>

        <div className={ui.reading}>
          <article className={styles.article}>

        <h2 style={H2}>Get help</h2>
        <p style={P}>
          Most questions are already answered: check the{" "}
          <Link href="/help">Help</Link> page for common problems, and the <Link href="/docs">documentation</Link> for how
          the platform works end to end.
        </p>
        <p style={P}>
          If you still need a human, use the support portal. The link is a placeholder for now and will be replaced
          once a real channel exists:
        </p>
        <p style={P}>
          <a href={SUPPORT_URL} target="_blank" rel="noreferrer">
            {SUPPORT_URL}
          </a>
        </p>
        <p style={{ ...P, color: "#8a8a8a", fontSize: "10.5pt" }}>
          Placeholder — replacing it only requires editing one constant in this page.
        </p>

        <h2 style={H2}>Report a bug or request a feature</h2>
        <p style={P}>
          The source code lives on GitHub. Issues and pull requests are the preferred way to report problems, because
          they keep the discussion in one place:
        </p>
        <p style={P}>
          <a href={GITHUB_URL} target="_blank" rel="noreferrer">
            {GITHUB_URL}
          </a>
        </p>
        <p style={P}>When opening an issue, include:</p>
        <ul style={UL}>
          <li style={LI}>What you expected to happen and what actually happened.</li>
          <li style={LI}>The project name and its status shown on the project page.</li>
          <li style={LI}>The relevant lines from the deployment log, if the build failed.</li>
          <li style={LI}>The failing request and status code from your browser Network tab.</li>
        </ul>

        <h2 style={H2}>Other ways to help</h2>
        <ul style={UL}>
          <li style={LI}>Star the repository or share it with someone who needs a simple static host.</li>
          <li style={LI}>Improve the documentation when you find a confusing sentence — small pull requests are welcome.</li>
          <li style={LI}>
            Report security issues privately instead of opening a public issue, and avoid testing against other
            people&apos;s projects.
          </li>
        </ul>

        <h2 style={H2}>License</h2>
        <p style={P}>
          only4.fun is released under the <strong>MIT License</strong>. You are free to use, copy, modify, merge,
          publish, distribute, sublicense, and sell copies of the software, provided the copyright notice and this
          permission notice are included in all copies or substantial portions of the software.
        </p>
        <p style={P}>
          The software is provided &quot;as is&quot;, without warranty of any kind, express or implied, including but
          not limited to the warranties of merchantability, fitness for a particular purpose and noninfringement. In no
          event shall the authors or copyright holders be liable for any claim, damages or other liability arising from,
          out of, or in connection with the software or the use of the software.
        </p>
        <p style={P}>
          Full text: <a href={LICENSE_URL} target="_blank" rel="noreferrer">MIT License</a>
          {" · "}
          <span style={{ fontSize: "11pt", color: "#8a8a8a" }}>
            Copyright (c) 2026 only4.fun contributors
          </span>
        </p>

        <h3 style={H3}>Third-party notices</h3>
        <p style={{ ...P, marginBottom: 0 }}>
          The dashboard is built with <a href="https://nextjs.org" target="_blank" rel="noreferrer">Next.js</a> and{" "}
          <a href="https://react.dev" target="_blank" rel="noreferrer">React</a>; the API uses{" "}
          <a href="https://gin-gonic.com" target="_blank" rel="noreferrer">Gin</a> and{" "}
          <a href="https://www.sqlite.org" target="_blank" rel="noreferrer">SQLite</a>. Each dependency keeps its own
          license, distributed with the packages. Build runs directly on the host using Node.js and npm. Sessions rely on
          GitHub OAuth, and subdomains are served through Cloudflare.
        </p>
        <p style={{ ...P, color: "#8a8a8a", marginTop: 24 }}>
          Code snippets in the documentation are examples only and carry no warranty. Product names and logos remain
          the property of their respective owners.
        </p>
          </article>
        </div>
      </main>
    </AppShell>
  )
}
