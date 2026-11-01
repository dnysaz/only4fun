import type { MetadataRoute } from "next"
import { SITE_URL } from "@/lib/site"

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: "/",
        // Halaman di balik login dan endpoint internal tidak berguna di indeks.
        disallow: ["/dashboard", "/projects", "/profile", "/login", "/api/", "/auth/"],
      },
      {
        // Bot AI tetap boleh membaca dokumentasi publik.
        userAgent: ["GPTBot", "ClaudeBot", "PerplexityBot", "Google-Extended"],
        allow: ["/", "/docs", "/help", "/support", "/terms", "/privacy"],
        disallow: ["/dashboard", "/projects", "/profile", "/login", "/api/", "/auth/"],
      },
    ],
    sitemap: `${SITE_URL}/sitemap.xml`,
    host: SITE_URL,
  }
}
