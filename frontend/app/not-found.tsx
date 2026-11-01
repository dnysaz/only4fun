import Link from "next/link"
import type { Metadata } from "next"
import styles from "./not-found.module.css"

export const metadata: Metadata = {
  title: "Page not found",
  robots: { index: false, follow: false },
}

const links = [
  { href: "/", label: "Home" },
  { href: "/docs", label: "Documentation" },
  { href: "/help", label: "Help" },
  { href: "/support", label: "Support" },
]

export default function NotFound() {
  return (
    <main className={styles.page}>
      <div className={styles.grid} aria-hidden="true" />
      <div className={styles.content}>
        <Link href="/" className={styles.brand}>
          Only4fun
        </Link>
        <p className={styles.code}>404</p>
        <h1 className={styles.title}>This page does not exist.</h1>
        <p className={styles.lede}>
          The link may be broken, or the project that used to live here has been deleted. Everything below is still
          reachable without signing in.
        </p>
        <nav className={styles.links} aria-label="Useful pages">
          {links.map((link) => (
            <Link key={link.href} href={link.href} className={styles.link}>
              {link.label}
            </Link>
          ))}
        </nav>
      </div>
    </main>
  )
}
