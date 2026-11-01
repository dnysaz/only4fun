import type { Metadata } from "next"

export const metadata: Metadata = {
  title: 'Terms & Conditions',
  description:
    'only4.fun is an open-source project built for fun. Free static hosting with no warranty, no uptime guarantee, and full responsibility for published content on the repository owner.',
  alternates: { canonical: "/terms" },
  openGraph: {
    type: "website",
    url: "/terms",
    title: 'Terms & Conditions',
    description:
      'only4.fun is an open-source project built for fun. Free static hosting with no warranty, no uptime guarantee, and full responsibility for published content on the repository owner.',
  },
}

export default function TermsLayout({ children }: { children: React.ReactNode }) {
  return children
}
