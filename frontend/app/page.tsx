"use client"

import Link from "next/link"
import { useCurrentUser } from "@/lib/auth"
import styles from "./home.module.css"

const promises = [
  {
    title: "Free forever",
    text: "No trial, no credit card, no per-seat pricing. Deploy and stay online for free.",
  },
  {
    title: "No limits",
    text: "No bandwidth caps, no build-minute quotas, no hidden usage meters.",
  },
  {
    title: "No soft starts",
    text: "No waiting room, no queue, no cold-start throttling. Your site is served instantly.",
  },
  {
    title: "Unlimited projects",
    text: "Deploy every open-source repository you own — there is no project limit.",
  },
]

function Tick() {
  return (
    <svg className={styles.tick} width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={3} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      <path d="M20 6 9 17l-5-5" />
    </svg>
  )
}

export default function HomePage() {
  const { user, loaded } = useCurrentUser()

  return (
    <div className={styles.page}>
      <div className={styles.grid} aria-hidden="true" />
      <div className={styles.glow} aria-hidden="true" />

      <div className={styles.content}>
        <span className={styles.badge}>
          <span className={styles.badgeDot} aria-hidden="true" />
          100% free · no limits · no soft starts
        </span>

        <h1 className={styles.title}>
          Ship your site.
          <br />
          Fast, Free And Simple
        </h1>

        <p className={styles.subtitle}>
          only4.fun builds a static or serverless site straight from your Git repo and serves it on{" "}
          <strong>your-project.only4.fun</strong>. Log in with GitHub, deploy in seconds — and it stays{" "}
          <strong>completely free</strong>: no limits, no soft starts, no cap on the number of open-source
          projects you can host.
        </p>

        <div className={styles.actions}>
          {loaded && user ? (
            <Link href="/dashboard" className={styles.cta}>
              <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                <rect x="3" y="3" width="7" height="7" rx="1.5" />
                <rect x="14" y="3" width="7" height="7" rx="1.5" />
                <rect x="3" y="14" width="7" height="7" rx="1.5" />
                <rect x="14" y="14" width="7" height="7" rx="1.5" />
              </svg>
              Open Dashboard
            </Link>
          ) : (
            <a href="/auth/github/login" className={styles.cta}>
              <svg width="17" height="17" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.44 9.8 8.21 11.39.6.11.82-.26.82-.58 0-.29-.01-1.04-.02-2.04-3.34.73-4.04-1.61-4.04-1.61-.55-1.39-1.35-1.76-1.35-1.76-1.11-.76.08-.75.08-.75 1.23.09 1.88 1.28 1.88 1.28 1.09 1.87 2.87 1.34 3.57 1.03.11-.8.43-1.34.78-1.65-2.7-.31-5.54-1.36-5.54-6.02 0-1.32.47-2.4 1.24-3.24-.12-.31-.54-1.56.12-3.26 0 0 1.01-.33 3.3 1.25A11.5 11.5 0 0112 5.8c1 .004 2.01.14 2.91.41 2.28-1.58 3.29-1.25 3.29-1.25.66 1.7.24 2.95.12 3.26.78.85 1.23 1.92 1.23 3.24 0 4.68-2.85 5.71-5.56 5.98.44.38.82 1.14.82 2.3 0 1.66-.01 2.98-.01 3.38 0 .32.22.71.83.57A12.004 12.004 0 0024 12c0-6.63-5.37-12-12-12z" />
              </svg>
              Continue with GitHub
            </a>
          )}
          <Link href="/docs" className={styles.ctaGhost}>
            Read the docs
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
              <path d="M5 12h14M13 6l6 6-6 6" />
            </svg>
          </Link>
        </div>

        <div className={styles.features}>
          {promises.map((item) => (
            <div className={styles.feature} key={item.title}>
              <p className={styles.featureTitle}>
                <Tick />
                {item.title}
              </p>
              <p className={styles.featureText}>{item.text}</p>
            </div>
          ))}
        </div>
      </div>

      <footer className={styles.footerWrap}>
        <div className={styles.footer}>
          <span>*.only4.fun</span>
          <span aria-hidden="true">·</span>
          <Link href="/docs">Docs</Link>
          <span aria-hidden="true">·</span>
          <Link href="/help">Help</Link>
          <span aria-hidden="true">·</span>
          <Link href="/support">Support</Link>
          <span aria-hidden="true">·</span>
          <Link href="/terms">Terms &amp; Conditions</Link>
          <span aria-hidden="true">·</span>
          <Link href="/privacy">Privacy Policy</Link>
          <span aria-hidden="true">·</span>
          <span>MIT licensed</span>
        </div>
        <p className={styles.footerNote}>
          An open-source project built for fun. Free static hosting with no warranty and no uptime guarantee —
          whatever a hosted site shows is the responsibility of its repository owner.
        </p>
        <p className={styles.footerSupport}>
          Supported by{" "}
          <a href="https://moogo.dev" target="_blank" rel="noreferrer">
            moogo.dev
          </a>{" "}
          — A Lightweight Database for Serverless Applications
        </p>
      </footer>
    </div>
  )
}