import type { Metadata } from "next"

export const metadata: Metadata = {
  title: 'Support',
  description:
    'Get help with only4.fun: where to report a bug, request a feature, or ask about hosting, domains, and deployments for this open-source project.',
  alternates: { canonical: "/support" },
  openGraph: {
    type: "website",
    url: "/support",
    title: 'Support',
    description:
      'Get help with only4.fun: where to report a bug, request a feature, or ask about hosting, domains, and deployments for this open-source project.',
  },
}

export default function SupportLayout({ children }: { children: React.ReactNode }) {
  return children
}
