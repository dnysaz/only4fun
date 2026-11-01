"use client"

import { useParams } from "next/navigation"
import Link from "next/link"
import { docSections } from "../content"
import styles from "../docs.module.css"

export default function DocSectionPage() {
  const params = useParams<{ slug: string }>()
  const slug = params?.slug || ""
  const section = docSections[slug]

  if (!section) {
    return (
      <div style={{ padding: "40px 48px" }}>
        <h1 style={{ fontSize: "23pt", fontWeight: 700, margin: "0 0 8px" }}>Not found</h1>
        <p style={{ fontSize: "11.5pt", color: "#8a8a8a" }}>
          This documentation page doesn&apos;t exist.{" "}
          <Link href="/docs" style={{ color: "#2563eb" }}>Back to docs</Link>.
        </p>
      </div>
    )
  }

  return (
    <div className={styles.article}>
      {section.content}
      <div style={{ marginTop: 32, paddingTop: 20, borderTop: "1px solid #f0f0f0", display: "flex", justifyContent: "space-between", gap: 12, flexWrap: "wrap" }}>
        <Link href="/docs" style={{ fontSize: "10.5pt", color: "#8a8a8a", textDecoration: "none" }}>
          ← All docs
        </Link>
      </div>
    </div>
  )
}
