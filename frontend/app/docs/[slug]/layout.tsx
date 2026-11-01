import type { Metadata } from "next"
import { docMeta } from "../meta"
import DocsBreadcrumbs from "../breadcrumbs"
import { SITE_URL } from "@/lib/site"

export function generateMetadata({ params }: { params: { slug: string } }): Metadata {
  const meta = docMeta[params.slug]
  if (!meta) {
    return {
      title: "Page not found",
      robots: { index: false, follow: false },
    }
  }
  return {
    title: meta.title,
    description: meta.description,
    alternates: { canonical: `/docs/${params.slug}` },
    openGraph: {
      type: "article",
      url: `/docs/${params.slug}`,
      title: meta.title,
      description: meta.description,
    },
    twitter: {
      card: "summary_large_image",
      title: meta.title,
      description: meta.description,
    },
  }
}

export default function DocSectionLayout({
  children,
  params,
}: {
  children: React.ReactNode
  params: { slug: string }
}) {
  const meta = docMeta[params.slug]
  if (!meta) return children

  const article = {
    "@context": "https://schema.org",
    "@type": "TechArticle",
    headline: meta.title,
    description: meta.description,
    url: `${SITE_URL}/docs/${params.slug}`,
    inLanguage: "en",
    isPartOf: { "@type": "WebSite", name: "Only4fun", url: SITE_URL },
  }

  return (
    <>
      <DocsBreadcrumbs title={meta.title} slug={params.slug} />
      <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: JSON.stringify(article) }} />
      {children}
    </>
  )
}
