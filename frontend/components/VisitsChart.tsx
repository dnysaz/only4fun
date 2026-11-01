"use client"

import { useId, useMemo, useState } from "react"
import styles from "./VisitsChart.module.css"

export interface DailyVisit {
  day: string
  count: number
}

const W = 640
const H = 200
const PAD = { top: 18, right: 8, bottom: 10, left: 8 }
const PLOT_W = W - PAD.left - PAD.right
const PLOT_H = H - PAD.top - PAD.bottom
const BASELINE = H - PAD.bottom

interface Point {
  x: number
  y: number
}

/** Melengkapi hari yang kosong dengan 0 supaya sumbu waktu tetap kontinu. */
export function buildVisitSeries(visits: DailyVisit[], days = 14): DailyVisit[] {
  const counts = new Map(visits.map((v) => [v.day.slice(0, 10), v.count]))
  const now = new Date()
  const today = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), now.getUTCDate()))
  const series: DailyVisit[] = []

  for (let offset = days - 1; offset >= 0; offset--) {
    const d = new Date(today)
    d.setUTCDate(today.getUTCDate() - offset)
    const key = d.toISOString().slice(0, 10)
    series.push({ day: key, count: counts.get(key) ?? 0 })
  }

  return series
}

export function formatDayLabel(day: string) {
  const d = new Date(`${day}T00:00:00Z`)
  if (Number.isNaN(d.getTime())) return day
  return d.toLocaleDateString("en-GB", { day: "numeric", month: "short", timeZone: "UTC" })
}

function niceMax(value: number) {
  if (value <= 4) return 4
  const power = Math.pow(10, Math.floor(Math.log10(value)))
  const n = value / power
  const step = n <= 1 ? 1 : n <= 2 ? 2 : n <= 5 ? 5 : 10
  return step * power
}

/**
 * Kurva Catmull-Rom yang diubah menjadi cubic Bezier. Titik kendali di-clamp
 * ke area plot supaya lengkungan tetap halus tanpa keluar dari grafik.
 */
function smoothPath(points: Point[]) {
  if (points.length === 0) return ""
  if (points.length === 1) return `M ${points[0].x} ${points[0].y}`

  const clampY = (y: number) => Math.min(BASELINE, Math.max(PAD.top, y))
  let d = `M ${points[0].x} ${points[0].y}`

  for (let i = 0; i < points.length - 1; i++) {
    const p0 = points[i - 1] ?? points[i]
    const p1 = points[i]
    const p2 = points[i + 1]
    const p3 = points[i + 2] ?? p2

    const c1x = p1.x + (p2.x - p0.x) / 6
    const c1y = clampY(p1.y + (p2.y - p0.y) / 6)
    const c2x = p2.x - (p3.x - p1.x) / 6
    const c2y = clampY(p2.y - (p3.y - p1.y) / 6)

    d += ` C ${c1x} ${c1y}, ${c2x} ${c2y}, ${p2.x} ${p2.y}`
  }

  return d
}

export default function VisitsChart({ visits }: { visits: DailyVisit[] }) {
  const gradientId = useId().replace(/:/g, "")
  const [active, setActive] = useState<number | null>(null)

  const series = useMemo(() => buildVisitSeries(visits), [visits])
  const peak = useMemo(() => Math.max(1, ...series.map((v) => v.count)), [series])
  const max = niceMax(peak)

  const points = useMemo<Point[]>(
    () =>
      series.map((v, i) => ({
        x: PAD.left + (series.length === 1 ? PLOT_W / 2 : (i / (series.length - 1)) * PLOT_W),
        y: PAD.top + (1 - v.count / max) * PLOT_H,
      })),
    [series, max]
  )

  const line = useMemo(() => smoothPath(points), [points])
  const area = useMemo(() => {
    if (points.length === 0) return ""
    const last = points[points.length - 1]
    const first = points[0]
    return `${line} L ${last.x} ${BASELINE} L ${first.x} ${BASELINE} Z`
  }, [line, points])

  const handleMove = (event: React.MouseEvent<SVGSVGElement>) => {
    const rect = event.currentTarget.getBoundingClientRect()
    if (!rect.width) return
    const x = ((event.clientX - rect.left) / rect.width) * W
    const ratio = Math.min(1, Math.max(0, (x - PAD.left) / PLOT_W))
    setActive(Math.round(ratio * (series.length - 1)))
  }

  const activeItem = active !== null ? series[active] : null
  const activePoint = active !== null ? points[active] : null
  const gridLines = [BASELINE, PAD.top + PLOT_H / 2, PAD.top]

  return (
    <div className={styles.wrap}>
      <div className={styles.head}>
        <div className={styles.readout} aria-live="polite">
          {activeItem ? (
            <>
              <strong>{activeItem.count}</strong>
              <span> page views on {formatDayLabel(activeItem.day)}</span>
            </>
          ) : (
            <span>Last 14 days · hover the chart for daily detail</span>
          )}
        </div>
        <div className={styles.legend}>
          <span className={styles.swatch} aria-hidden="true" />
          page views
        </div>
      </div>

      <div className={styles.plot}>
        <span className={styles.yMax} aria-hidden="true">
          {max}
        </span>
        <svg
          className={styles.chart}
          viewBox={`0 0 ${W} ${H}`}
          preserveAspectRatio="none"
          role="img"
          aria-label="Daily page views for the last 14 days"
          onMouseMove={handleMove}
          onMouseLeave={() => setActive(null)}
        >
          <defs>
            <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stopColor="#2563eb" stopOpacity="0.28" />
              <stop offset="60%" stopColor="#2563eb" stopOpacity="0.08" />
              <stop offset="100%" stopColor="#2563eb" stopOpacity="0" />
            </linearGradient>
          </defs>

          {gridLines.map((y, i) => (
            <line
              key={i}
              x1={PAD.left}
              y1={y}
              x2={W - PAD.right}
              y2={y}
              className={styles.gridLine}
              vectorEffect="non-scaling-stroke"
            />
          ))}

          <path d={area} fill={`url(#${gradientId})`} />
          <path
            d={line}
            fill="none"
            stroke="#2563eb"
            strokeWidth={2.2}
            strokeLinecap="round"
            strokeLinejoin="round"
            vectorEffect="non-scaling-stroke"
          />

          {activePoint && (
            <>
              <line
                x1={activePoint.x}
                y1={PAD.top}
                x2={activePoint.x}
                y2={BASELINE}
                className={styles.cursor}
                vectorEffect="non-scaling-stroke"
              />
              <line
                x1={PAD.left}
                y1={activePoint.y}
                x2={activePoint.x}
                y2={activePoint.y}
                className={styles.cursor}
                vectorEffect="non-scaling-stroke"
              />
              <circle cx={activePoint.x} cy={activePoint.y} r={4} className={styles.point} vectorEffect="non-scaling-stroke" />
            </>
          )}
        </svg>
      </div>

      <div className={styles.xAxis} aria-hidden="true">
        {series.map((v) => (
          <span key={v.day} className={styles.xLabel} title={formatDayLabel(v.day)}>
            {Number(v.day.slice(8, 10))}
          </span>
        ))}
      </div>
    </div>
  )
}
