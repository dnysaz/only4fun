import type { Metadata } from "next"

export const metadata: Metadata = {
  title: 'Privacy Policy',
  description:
    'What only4.fun stores (GitHub account basics, an encrypted token, session records, project data, aggregate visit counts), which cookies it uses, and what it never does with your data.',
  alternates: { canonical: "/privacy" },
  openGraph: {
    type: "website",
    url: "/privacy",
    title: 'Privacy Policy',
    description:
      'What only4.fun stores (GitHub account basics, an encrypted token, session records, project data, aggregate visit counts), which cookies it uses, and what it never does with your data.',
  },
}

export default function PrivacyLayout({ children }: { children: React.ReactNode }) {
  return children
}
