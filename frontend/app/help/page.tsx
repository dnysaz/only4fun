"use client"

import Link from "next/link"
import AppShell from "@/components/AppShell"
import { useCurrentUser } from "@/lib/auth"
import ui from "../ui.module.css"
import styles from "../content.module.css"

const H2: React.CSSProperties = { fontSize: "16px", fontWeight: 650, margin: "34px 0 12px", paddingTop: 24, borderTop: "1px solid #e4e4e4", letterSpacing: "-0.1px" }
const H3: React.CSSProperties = { fontSize: "14px", fontWeight: 650, margin: "24px 0 6px" }
const P: React.CSSProperties = { fontSize: "13.5px", lineHeight: 1.75, color: "#4f4f4f", margin: "0 0 12px" }
const UL: React.CSSProperties = { margin: "0 0 12px", paddingLeft: 22 }
const LI: React.CSSProperties = { fontSize: "13.5px", lineHeight: 1.75, color: "#4f4f4f", marginBottom: 6 }
const CODE: React.CSSProperties = {
  background: "#f4f4f4",
  padding: "1px 5px",
  borderRadius: 4,
  fontSize: "12.5px",
  fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
}

const QUESTIONS: { q: string; a: React.ReactNode }[] = [
  {
    q: "The dashboard asks me to sign in even though I just did",
    a: (
      <>
        Sessions expire after 24 hours and are invalid when the server's encryption key changes. Reload with a hard
        refresh, then sign in once more. If a fresh sign-in still bounces you back, open{" "}
        <code style={CODE}>/health</code> on the API host: if it does not answer, the API server is down and the
        dashboard has no session to read.
      </>
    ),
  },
  {
    q: "Login sends me to a page that cannot be reached",
    a: (
      <>
        The GitHub callback always returns to the API server first, so it must be running and reachable before you
        click the button. Start the backend, confirm <code style={CODE}>/health</code> answers, then retry.
      </>
    ),
  },
  {
    q: "My repository is not in the repository list",
    a: (
      <>
        Only the 100 most recently updated repositories are requested from GitHub. Push a commit to the repository, or
        re-authorize the app so it can read your repository list again. If the list is empty with an error, the GitHub
        token has likely been revoked. In a hurry? Switch to Upload ZIP on the New project page and skip GitHub
        entirely.
      </>
    ),
  },
  {
    q: "Deploy fails, or the button seems to do nothing",
    a: (
      <>
        The form shows the error returned by the API. A duplicate project name and an invalid repository URL are the
        two most common causes. Project names accept lowercase letters, numbers, and dashes only, and must not start or
        end with a dash.
      </>
    ),
  },
  {
    q: "The project is stuck on pending",
    a: (
      <>
        Pending means the deployment is queued but no build has started. Builds run one at a time by default, so a
        deployment can wait behind another. If nothing changes for more than ten minutes, open the project page and
        read the deployment log.
      </>
    ),
  },
  {
    q: "The build failed while installing dependencies",
    a: (
      <>
        The log shows the exact command and its output. Usual causes: a lockfile that does not match the manifest, a
        dependency that needs a private registry credential, or a Node version requirement newer than the build image.
      </>
    ),
  },
  {
    q: "The subdomain shows a 404 or the wrong page",
    a: (
      <>
        The build must write an <code style={CODE}>index.html</code> into one of the detected output folders:{" "}
        <code style={CODE}>dist</code>, <code style={CODE}>out</code>, <code style={CODE}>build</code>,{" "}
        <code style={CODE}>public</code>, or the repository root. Confirm the status is deployed and the log ends with
        a publish step.
      </>
    ),
  },
  {
    q: "I changed an environment variable but the site is unchanged",
    a: (
      <>
        Variables are applied at build time. Save them, then click Redeploy so the build runs again with the new
        values. An already published site is not modified by editing variables.
      </>
    ),
  },
  {
    q: "Where can I see visits and build logs?",
    a: (
      <>
        On the project detail page, opened by clicking a project card on the dashboard. It shows total page views with
        a daily chart, repository metadata, environment variables, and every deployment with its log. Only page requests
        are counted; assets such as CSS and images are not.
      </>
    ),
  },
  {
    q: "How do I sign out?",
    a: (
      <>
        Use Logout at the bottom of the sidebar (on a phone, open the menu with the ☰ button first), or open Profile.
        That deletes the session on the server and clears the cookie in this browser.
      </>
    ),
  },
  {
    q: "Does pushing to GitHub redeploy automatically?",
    a: (
      <>
        Yes, if you connect the webhook shown on the project page (Auto-deploy webhook). Add it in GitHub under
        Settings → Webhooks with the given payload URL and secret; pushes to the default branch then start a build
        automatically. Without a webhook, deploy manually with Redeploy.
      </>
    ),
  },
  {
    q: "Can I deploy without a GitHub repository?",
    a: (
      <>
        Yes. On New project, switch to Upload ZIP and drop an archive containing index.html at its root (or inside one
        top-level folder). ZIP deployments are static only and cannot be redeployed from a repository; upload a new
        project to publish a new version.
      </>
    ),
  },
  {
    q: "Can I use my own domain?",
    a: (
      <>
        Yes. Open the project, use Custom domain, and enter the hostname. Then add a CNAME record for that hostname
        pointing to your project subdomain. TLS is issued by Cloudflare; if automatic issuance is unavailable, serve
        the domain through your own Cloudflare zone.
      </>
    ),
  },
]

export default function HelpPage() {
  // Help bisa dibaca tanpa login: kalau sesi tidak ada, cukup jadi tamu.
  // Cache dari halaman sebelumnya dipakai agar shell tidak berkedip.
  const { user, loaded } = useCurrentUser()

  return (
    <AppShell user={user} loading={!loaded} publicShell>
      <main className={`${ui.page} ${ui.pageWide}`}>
        <div className={ui.topbar}>
          <Link href="/dashboard" className={ui.back}>
            ← Dashboard
          </Link>
          <span className={ui.eyebrow}>Help</span>
        </div>
        <header className={ui.head}>
          <div className={ui.headMain}>
            <h1 className={ui.title}>Help</h1>
            <p className={ui.lede}>
              Short answers to the problems people hit most often. For the full explanation of how the platform works,
              read the <Link href="/docs">documentation</Link>.
            </p>
          </div>
        </header>

        <div className={ui.reading}>
          <article className={styles.article}>

        <h2 style={H2}>Three checks that solve most problems</h2>
        <ol style={UL}>
          <li style={LI}>
            Is the API alive? Open <code style={CODE}>/health</code> on the API host. It should return{" "}
            <code style={CODE}>{'{"status":"ok"}'}</code>.
          </li>
          <li style={LI}>Hard refresh the page, then sign in with GitHub again.</li>
          <li style={LI}>
            Open the browser DevTools Network tab and look at the failing request: its status code and response body
            usually name the cause.
          </li>
        </ol>

        <h2 style={H2}>Common questions</h2>
        {QUESTIONS.map((item) => (
          <div key={item.q} style={{ marginBottom: 22 }}>
            <h3 style={H3}>{item.q}</h3>
            <p style={{ ...P, marginBottom: 0 }}>{item.a}</p>
          </div>
        ))}

        <h2 style={H2}>Still stuck</h2>
        <p style={P}>
          Collect three things before asking for help: the project name, the status shown on the project page, and the
          failing request from the Network tab. With those, almost any problem can be reproduced. Send them to whoever
          operates this only4.fun instance.
        </p>
        <p style={{ ...P, marginBottom: 0 }}>
          <Link href="/docs">Read the documentation</Link>
          {" · "}
          <Link href="/projects/new">Create a project</Link>
        </p>
          </article>
        </div>
      </main>
    </AppShell>
  )
}
