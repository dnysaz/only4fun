import type { Metadata } from "next"
import DocsShell from "@/components/DocsShell"
import DocsBreadcrumbs from "./breadcrumbs"
import { docMeta } from "./meta"

// Layout ini sengaja server component agar bisa mengekspor metadata; bagian
// interaktif (dashboard shell, pencarian, panel mobile) ada di DocsShell.
export const metadata: Metadata = {
  title: docMeta[""].title,
  description: docMeta[""].description,
  alternates: { canonical: "/docs" },
  openGraph: {
    type: "website",
    url: "/docs",
    title: docMeta[""].title,
    description: docMeta[""].description,
  },
}

export default function DocsLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <DocsBreadcrumbs title="Documentation" />
      <DocsShell>{children}</DocsShell>
    </>
  )
}
