import { SITE_URL } from "@/lib/site"

/**
 * BreadcrumbList untuk mesin pencari: Home › Docs › Artikel.
 * Dirender sebagai script JSON-LD (tanpa tampilan) di server component.
 */
export default function DocsBreadcrumbs({ title, slug }: { title: string; slug?: string }) {
  const items = [
    { name: "Home", url: `${SITE_URL}/` },
    { name: "Documentation", url: `${SITE_URL}/docs` },
  ]
  if (slug) {
    items.push({ name: title, url: `${SITE_URL}/docs/${slug}` })
  }

  const data = {
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    itemListElement: items.map((item, index) => ({
      "@type": "ListItem",
      position: index + 1,
      name: item.name,
      item: item.url,
    })),
  }

  return <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(data) }} />
}
