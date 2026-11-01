"use client"

import { useEffect, useRef, useState } from "react"
import styles from "./BuildTerminal.module.css"

export interface TerminalDeployment {
  id: string
  status: string
  log: string
  created_at: string
  updated_at: string
}

function formatDuration(ms: number) {
  const total = Math.max(0, Math.floor(ms / 1000))
  const minutes = Math.floor(total / 60)
  const seconds = total % 60
  return minutes > 0 ? `${minutes}m ${String(seconds).padStart(2, "0")}s` : `${seconds}s`
}

function statusInfo(projectStatus: string) {
  if (projectStatus === "building") {
    return { label: "Building", tone: styles.live }
  }
  return { label: "Queued", tone: styles.queued }
}

export default function BuildTerminal({
  projectStatus,
  deployment,
}: {
  projectStatus: string
  deployment?: TerminalDeployment | null
}) {
  const [now, setNow] = useState(() => Date.now())
  const bodyRef = useRef<HTMLPreElement>(null)

  // Timer for the duration of the running build.
  useEffect(() => {
    setNow(Date.now())
    const timer = setInterval(() => setNow(Date.now()), 1000)
    return () => clearInterval(timer)
  }, [])

  const log = deployment?.log || ""

  // Always follow the latest output while the build is running.
  useEffect(() => {
    const el = bodyRef.current
    if (el) el.scrollTop = el.scrollHeight
  }, [log])

  const status = statusInfo(projectStatus)
  const started = deployment?.created_at ? new Date(deployment.created_at).getTime() : NaN
  const elapsed = projectStatus === "building" && !Number.isNaN(started) ? formatDuration(now - started) : ""
  const meta = elapsed || "waiting for a worker"

  return (
    <section className={styles.terminal} aria-label="Build terminal">
      <header className={styles.bar}>
        <span className={styles.title}>Build terminal</span>
        <span className={`${styles.status} ${status.tone}`}>
          <span className={styles.statusDot} aria-hidden="true" />
          {status.label}
        </span>
        <span className={styles.spacer} />
        <span className={styles.meta}>{meta}</span>
      </header>

      <pre ref={bodyRef} className={styles.body}>
        <span className={styles.prompt}>$</span> build {deployment ? deployment.id.slice(0, 8) : "queued"}
        {"\n"}
        {log || "waiting for build output..."}
        <span className={styles.cursor} aria-hidden="true" />
      </pre>
    </section>
  )
}
