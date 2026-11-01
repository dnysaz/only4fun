import type { Metadata } from "next"

export const metadata: Metadata = {
  title: 'Help',
  description:
    'Answers to the questions that come up most often on only4.fun: sign-in problems, failed builds, missing index.html, custom domains, webhooks, and how to delete a project.',
  alternates: { canonical: "/help" },
  openGraph: {
    type: "website",
    url: "/help",
    title: 'Help',
    description:
      'Answers to the questions that come up most often on only4.fun: sign-in problems, failed builds, missing index.html, custom domains, webhooks, and how to delete a project.',
  },
}

export default function HelpLayout({ children }: { children: React.ReactNode }) {
  return children
}
