"use client"

import Link from "next/link"
import { docSections } from "./content"
import styles from "./docs.module.css"

export default function DocsOverviewPage() {
  const overview = docSections[""]

  return (
    <div className={styles.article}>
      {overview.content}
      <div style={{ marginTop: 32, paddingTop: 20, borderTop: "1px solid #f0f0f0" }}>
        <h2 style={{ fontSize: "14pt", fontWeight: 600, margin: "0 0 12px" }}>Quick links</h2>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(200px, 1fr))", gap: 8 }}>
          {Object.entries(docSections)
            .filter(([slug]) => slug !== "")
            .slice(0, 8)
            .map(([slug, section]) => (
              <Link
                key={slug}
                href={`/docs/${slug}`}
                style={{
                  display: "block",
                  padding: "10px 12px",
                  borderRadius: 8,
                  border: "1px solid #eaeaea",
                  fontSize: "10.5pt",
                  color: "#3d3d3d",
                  textDecoration: "none",
                  transition: "border-color 0.15s",
                }}
              >
                {section.title}
              </Link>
            ))}
        </div>
      </div>
    </div>
  )
}
