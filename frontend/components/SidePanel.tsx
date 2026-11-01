"use client"

import { useEffect, useRef } from "react"
import styles from "./SidePanel.module.css"

interface SidePanelProps {
  open: boolean
  title: string
  subtitle?: string
  onClose: () => void
  children: React.ReactNode
}

// SidePanel menampilkan isi dalam panel yang meluncur masuk dari kanan,
// dengan latar gelap tipis di belakangnya.
export default function SidePanel({ open, title, subtitle, onClose, children }: SidePanelProps) {
  const closeRef = useRef<HTMLButtonElement>(null)
  // Simpan callback terbaru tanpa memicu ulang effect saat parent re-render,
  // supaya fokus dan kunci scroll tidak di-reset setiap kali state berubah.
  const onCloseRef = useRef(onClose)
  onCloseRef.current = onClose

  useEffect(() => {
    if (!open) return

    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onCloseRef.current()
    }
    document.addEventListener("keydown", onKeyDown)

    // Kunci scroll halaman selama panel terbuka.
    const previousOverflow = document.body.style.overflow
    document.body.style.overflow = "hidden"
    closeRef.current?.focus()

    return () => {
      document.removeEventListener("keydown", onKeyDown)
      document.body.style.overflow = previousOverflow
    }
  }, [open])

  if (!open) return null

  return (
    <div className={styles.root}>
      <div className={styles.backdrop} onClick={onClose} aria-hidden="true" />
      <aside className={styles.panel} role="dialog" aria-modal="true" aria-label={title}>
        <header className={styles.head}>
          <div className={styles.headMain}>
            <h2 className={styles.title}>{title}</h2>
            {subtitle && <p className={styles.subtitle}>{subtitle}</p>}
          </div>
          <button ref={closeRef} type="button" className={styles.close} onClick={onClose} aria-label="Close panel">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </header>
        <div className={styles.body}>{children}</div>
      </aside>
    </div>
  )
}
