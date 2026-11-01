"use client"

import { useMemo } from "react"
import styles from "./ContributionGraph.module.css"

export interface ContributionDay {
  date: string
  level: number
  count: number
}

function dayTitle(day: ContributionDay) {
  if (day.count > 0) {
    return `${day.count} contribution${day.count === 1 ? "" : "s"} on ${day.date}`
  }
  if (day.count === 0) return `No contributions on ${day.date}`
  return day.date
}

const LEGEND_LEVELS = [0, 1, 2, 3, 4]
const DAY_MS = 24 * 60 * 60 * 1000

function parseDate(value: string) {
  // Tanggal dari GitHub berbentuk YYYY-MM-DD; dihitung sebagai UTC agar tidak
  // bergeser satu hari karena zona waktu lokal.
  const [year, month, day] = value.split("-").map(Number)
  if (!year || !month || !day) return null
  return Date.UTC(year, month - 1, day)
}

interface Cell {
  day: ContributionDay
  column: number
  row: number
}

/**
 * Kalender kontribusi ala GitHub. Posisi setiap kotak dihitung dari tanggalnya
 * (kolom = minggu, baris = hari dalam minggu), bukan dari urutan markup GitHub,
 * sehingga tampilan tetap benar walau susunan HTML berubah atau minggu pertama
 * tidak penuh.
 */
export default function ContributionGraph({ days, total }: { days: ContributionDay[]; total: number }) {
  const layout = useMemo(() => {
    const dated = days
      .map((day) => ({ day, time: parseDate(day.date) }))
      .filter((entry): entry is { day: ContributionDay; time: number } => entry.time !== null)
      .sort((a, b) => a.time - b.time)

    if (dated.length === 0) return { cells: [] as Cell[], columns: 0 }

    const start = dated[0].time
    // Hari pertama graf GitHub hampir selalu Minggu; offset dihitung dari
    // tanggalnya sendiri agar tetap tepat bila bukan Minggu.
    const startWeekday = new Date(start).getUTCDay()

    const cells: Cell[] = dated.map(({ day, time }) => {
      const offset = Math.round((time - start) / DAY_MS)
      const absolute = offset + startWeekday
      return {
        day,
        column: Math.floor(absolute / 7),
        row: absolute % 7,
      }
    })

    const columns = Math.max(...cells.map((cell) => cell.column)) + 1
    return { cells, columns }
  }, [days])

  if (days.length === 0) {
    return <p className={styles.empty}>No contribution calendar is available for this account.</p>
  }

  return (
    <div className={styles.wrap}>
      <p className={styles.total}>
        <strong>{total.toLocaleString("en-GB")}</strong> contributions in the last year
      </p>

      <div className={styles.scroller}>
        <div
          className={styles.grid}
          style={{ gridTemplateColumns: `repeat(${Math.max(1, layout.columns)}, 10px)` }}
          role="img"
          aria-label={`Contribution calendar: ${total} contributions in the last year`}
        >
          {layout.cells.map((cell) => (
            <span
              key={cell.day.date}
              className={styles.cell}
              data-level={Math.min(4, Math.max(0, cell.day.level))}
              title={dayTitle(cell.day)}
              style={{ gridColumn: cell.column + 1, gridRow: cell.row + 1 }}
            />
          ))}
        </div>
      </div>

      <div className={styles.legend}>
        <span>Less</span>
        {LEGEND_LEVELS.map((level) => (
          <span key={level} className={styles.cell} data-level={level} aria-hidden="true" />
        ))}
        <span>More</span>
      </div>
    </div>
  )
}
