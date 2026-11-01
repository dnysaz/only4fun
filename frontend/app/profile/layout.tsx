import type { Metadata } from "next"

// Halaman di balik login: tidak perlu diindeks mesin pencari.
export const metadata: Metadata = {
  robots: { index: false, follow: false },
}

export default function ProfileLayout({ children }: { children: React.ReactNode }) {
  return children
}
