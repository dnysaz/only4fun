"use client"

import Link from "next/link"
import AppShell from "@/components/AppShell"
import { useCurrentUser } from "@/lib/auth"
import ui from "../ui.module.css"
import styles from "../content.module.css"

const H2: React.CSSProperties = { fontSize: "16px", fontWeight: 650, margin: "34px 0 12px", paddingTop: 24, borderTop: "1px solid #e4e4e4", letterSpacing: "-0.1px" }
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

export default function PrivacyPage() {
  const { user, loaded } = useCurrentUser()

  return (
    <AppShell user={user} loading={!loaded} publicShell>
      <main className={`${ui.page} ${ui.pageWide}`}>
        <div className={ui.topbar}>
          <Link href="/" className={ui.back}>
            ← Home
          </Link>
          <span className={ui.eyebrow}>Legal</span>
        </div>

        <header className={ui.head}>
          <div className={ui.headMain}>
            <h1 className={ui.title}>Privacy Policy.</h1>
            <p className={ui.lede}>
              What this open-source project stores, why it stores it, and what it never does with your data.
            </p>
          </div>
        </header>

        <div className={`${ui.reading} ${styles.article}`}>
          <p style={P}>
            only4.fun is a free, open-source hosting project for static sites. The data it keeps is the minimum needed
            to sign you in, build your repository, and serve the result on a subdomain.
          </p>

          <h2 style={H2}>1. What we store</h2>
          <ul style={UL}>
            <li style={LI}>
              <strong>GitHub account basics</strong> — numeric GitHub ID, username, email address, and avatar URL,
              received during sign-in.
            </li>
            <li style={LI}>
              <strong>GitHub access token</strong> — used to list, clone, and read your repositories. It is encrypted
              at rest with AES-256-GCM.
            </li>
            <li style={LI}>
              <strong>Session record</strong> — a random session identifier plus its SHA-256 hash and expiry time
              (24 hours).
            </li>
            <li style={LI}>
              <strong>Project data</strong> — project name/subdomain, repository URL or ZIP source, build status,
              deployment logs, environment variable names and values (values encrypted), webhook secret (encrypted),
              and custom domain.
            </li>
            <li style={LI}>
              <strong>Visit counters</strong> — an aggregate number of page views per project per day. No IP addresses,
              no cookies, and no visitor profiles are attached to those counters.
            </li>
          </ul>

          <h2 style={H2}>2. Cookies</h2>
          <ul style={UL}>
            <li style={LI}>
              <code style={CODE}>session_id</code> — <code style={CODE}>HttpOnly</code> cookie that keeps you signed in
              for 24 hours.
            </li>
            <li style={LI}>
              <code style={CODE}>oauth_state</code> — short-lived cookie that protects the GitHub sign-in flow against
              CSRF.
            </li>
            <li style={LI}>
              <code style={CODE}>oauth_next</code> — short-lived cookie remembering which page to return to after
              sign-in; deleted as soon as it is read.
            </li>
          </ul>
          <p style={P}>
            There are no advertising cookies, no tracking pixels, and no third-party analytics scripts on this site.
          </p>

          <h2 style={H2}>3. What we do not do</h2>
          <ul style={UL}>
            <li style={LI}>We do not sell, rent, or share your data with advertisers or data brokers.</li>
            <li style={LI}>We do not use your repositories for anything other than building and serving your projects.</li>
            <li style={LI}>We do not send marketing email.</li>
            <li style={LI}>
              We do not read the content of the sites you host; whatever a hosted page displays is the responsibility of
              the repository owner.
            </li>
          </ul>

          <h2 style={H2}>4. Third parties</h2>
          <p style={P}>
            Sign-in and repository access go through <strong>GitHub</strong>, so GitHub&apos;s own privacy policy applies
            to that part of the flow. If a custom domain or DNS record is created, the configured DNS provider (for
            example Cloudflare) processes the DNS data needed to publish it.
          </p>

          <h2 style={H2}>5. Data retention and deletion</h2>
          <ul style={UL}>
            <li style={LI}>Sessions are deleted when they expire or when you sign out.</li>
            <li style={LI}>Deleting a project removes its deployment records, release files, and DNS record.</li>
            <li style={LI}>
              Visit counters and deployment logs are kept while the project exists, so the owner can see traffic and
              build history.
            </li>
            <li style={LI}>
              Because this is a self-hostable project, whoever runs the server holds the database. On a public instance,
              ask the operator for deletion of your account data.
            </li>
          </ul>

          <h2 style={H2}>6. Security</h2>
          <p style={P}>
            Sessions are random 256-bit identifiers stored only as hashes. GitHub tokens and environment values are
            encrypted with AES-256-GCM. Projects are isolated per user, and only the owner can read or change them.
            That said — this is an open-source project made for fun, run without a commercial security guarantee, so do
            not publish secrets that would harm you if they leaked.
          </p>

          <h2 style={H2}>7. No warranty</h2>
          <p style={P}>
            As described in the <Link href="/terms">Terms &amp; Conditions</Link>, the service is provided as is, with
            no uptime guarantee and no warranty of any kind.
          </p>

          <h2 style={H2}>8. Contact</h2>
          <p style={P}>
            For privacy questions, use the <Link href="/support">Support page</Link> or open an issue in the project
            repository.
          </p>
        </div>
      </main>
    </AppShell>
  )
}
