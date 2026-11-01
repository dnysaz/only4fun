import { ImageResponse } from "next/og"

export const runtime = "edge"
export const alt = "Only4fun — ship your site, fast, free and simple"
export const size = { width: 1200, height: 630 }
export const contentType = "image/png"

// Gambar pratinjau saat tautan dibagikan di X, WhatsApp, Slack, Discord, dll.
export default async function OpengraphImage() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          padding: "80px",
          background: "#ffffff",
          backgroundImage:
            "linear-gradient(to right, rgba(0,0,0,0.06) 1px, transparent 1px), linear-gradient(to bottom, rgba(0,0,0,0.06) 1px, transparent 1px)",
          backgroundSize: "64px 64px",
          fontFamily: "sans-serif",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 16, marginBottom: 28 }}>
          <div
            style={{
              width: 56,
              height: 56,
              borderRadius: 14,
              background: "#111111",
              color: "#ffffff",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              fontSize: 30,
              fontWeight: 700,
            }}
          >
            4
          </div>
          <div style={{ fontSize: 26, fontWeight: 600, color: "#8a8a8a", letterSpacing: 0.5 }}>Only4fun</div>
        </div>

        <div
          style={{
            fontSize: 78,
            fontWeight: 800,
            letterSpacing: -2.5,
            lineHeight: 1.05,
            color: "#111111",
            display: "flex",
            flexDirection: "column",
          }}
        >
          <span>Ship your site.</span>
          <span style={{ color: "#6f6f6f" }}>Fast, Free And Simple</span>
        </div>

        <div style={{ marginTop: 30, fontSize: 30, lineHeight: 1.45, color: "#5c5c5c", maxWidth: 940 }}>
          Deploy a static or serverless site straight from your Git repo to your-project.only4.fun — no limits, no
          soft starts, no credit card.
        </div>

        <div style={{ marginTop: 36, display: "flex", alignItems: "center", gap: 12, fontSize: 24, color: "#111111" }}>
          <span
            style={{
              width: 12,
              height: 12,
              borderRadius: 999,
              background: "#22c55e",
              display: "flex",
            }}
          />
          <span>only4.fun</span>
          <span style={{ color: "#c4c4c4" }}>·</span>
          <span style={{ color: "#8a8a8a" }}>open source · free forever</span>
        </div>
      </div>
    ),
    size,
  )
}
