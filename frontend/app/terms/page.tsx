"use client"

import Link from "next/link"
import AppShell from "@/components/AppShell"
import { useCurrentUser } from "@/lib/auth"
import ui from "../ui.module.css"
import styles from "../content.module.css"

const H2: React.CSSProperties = { fontSize: "16px", fontWeight: 650, margin: "34px 0 12px", paddingTop: 24, borderTop: "1px solid #e4e4e4", letterSpacing: "-0.1px" }
const P: React.CSSProperties = { fontSize: "13.5px", lineHeight: 1.75, color: "#4f4f4f", margin: "0 0 12px" }
const UL: React.CSSProperties = { margin: "0 0 12px", paddingLeft: 22 }
const LI: React.CSSProperties = { fontSize: "13.5px", lineHeight: 1.75, color: "#4f4f4f", marginBottom: 6 }
const CODE: React.CSSProperties = {
  background: "#f4f4f4",
  padding: "1px 5px",
  borderRadius: 4,
  fontSize: "12.5px",
  fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
}

export default function TermsPage() {
  const { user, loaded } = useCurrentUser()

  return (
    <AppShell user={user} loading={!loaded} publicShell>
      <main className={`${ui.page} ${ui.pageWide}`}>
        <div className={ui.topbar}>
          <Link href="/" className={ui.back}>
            ← Home
          </Link>
          <span className={ui.eyebrow}>Legal</span>
        </div>

        <header className={ui.head}>
          <div className={ui.headMain}>
            <h1 className={ui.title}>Terms &amp; Conditions.</h1>
            <p className={ui.lede}>
              The short version: only4.fun is an open-source hobby project that gives your static site a free
              subdomain. It comes with no promises, no warranty, and no uptime guarantee.
            </p>
          </div>
        </header>

        <div className={`${ui.reading} ${styles.article}`}>
          <p style={P}>
            By creating an account, deploying a project, or visiting a site served on a <code style={CODE}>*.only4.fun</code>{" "}
            subdomain, you agree to the terms below. If you do not agree, please do not use the service.
          </p>

          <h2 style={H2}>1. A free, open-source project built for fun</h2>
          <p style={P}>
            only4.fun is an <strong>open-source project made for fun</strong>. It exists so developers can host a
            static site for free, learn how a deployment pipeline works, and share something small on the internet.
            There is no company behind it, no paid plan, and no commercial support contract.
          </p>

          <h2 style={H2}>2. What the service provides</h2>
          <p style={P}>
            only4.fun provides <strong>subdomain hosting only</strong>: it builds your repository (or accepts a ZIP
            archive) and serves the resulting files on a subdomain such as{" "}
            <code style={CODE}>your-project.only4.fun</code>, optionally on a domain you point at the platform.
          </p>
          <ul style={UL}>
            <li style={LI}>We do not sell, review, edit, or endorse the content published under a subdomain.</li>
            <li style={LI}>We do not provide email, databases, backend processing, or any service beyond serving files.</li>
            <li style={LI}>Features may be added, changed, or removed at any time without notice.</li>
          </ul>

          <h2 style={H2}>3. No warranty, no uptime guarantee</h2>
          <p style={P}>
            The service is provided <strong>&quot;as is&quot; and &quot;as available&quot;</strong>, without warranty of any
            kind, express or implied, including merchantability, fitness for a particular purpose, and
            non-infringement.
          </p>
          <ul style={UL}>
            <li style={LI}>
              There is <strong>no guarantee that the service stays online 24 hours a day, 7 days a week</strong>, and no
              uptime percentage is promised.
            </li>
            <li style={LI}>
              Deployments, builds, custom domains, and the platform itself may fail, be slow, or be unavailable at any
              time — including while you are publishing or presenting your site.
            </li>
            <li style={LI}>
              Projects may be removed, suspended, or the entire service may be shut down without prior notice. Keep
              your own copy of everything you deploy.
            </li>
          </ul>

          <h2 style={H2}>4. Your content is your responsibility</h2>
          <p style={P}>
            <strong>All information presented on a hosted site is the sole responsibility of the repository owner</strong>{" "}
            (or the person who uploaded the archive). That includes text, images, code, claims, prices, legal notices,
            links, and any data shown to visitors. only4.fun only provides the hosting of the subdomain and does not
            verify the accuracy, legality, or safety of that content.
          </p>
          <p style={P}>By deploying, you confirm that:</p>
          <ul style={UL}>
            <li style={LI}>You own the content or have the rights and licenses needed to publish it.</li>
            <li style={LI}>Your site does not break the law, infringe rights, distribute malware, phish, or spam.</li>
            <li style={LI}>You are responsible for any consequence of publishing it, including claims from third parties.</li>
          </ul>

          <h2 style={H2}>5. Acceptable use</h2>
          <ul style={UL}>
            <li style={LI}>No illegal, abusive, hateful, or harmful content, and no malware or phishing.</li>
            <li style={LI}>No attempts to break, overload, or probe the platform or other users&apos; projects.</li>
            <li style={LI}>No using the platform as a file dump, CDN for unrelated assets, or proxy for other services.</li>
          </ul>
          <p style={P}>
            Projects that break these rules may be deleted at any time, with or without notice, and the account may be
            blocked from deploying again.
          </p>

          <h2 style={H2}>6. Accounts and availability</h2>
          <p style={P}>
            Accounts are created through GitHub sign-in. You are responsible for keeping your GitHub account secure,
            because access to this service follows that account. Sessions expire after 24 hours.
          </p>

          <h2 style={H2}>7. Limitation of liability</h2>
          <p style={P}>
            To the maximum extent permitted by law, the maintainers and contributors of only4.fun are not liable for any
            direct, indirect, incidental, or consequential damage — including lost data, lost traffic, lost revenue, or
            reputational harm — arising from using or being unable to use the service.
          </p>

          <h2 style={H2}>8. Changes to these terms</h2>
          <p style={P}>
            These terms may be updated as the project evolves. The current version is always the one published on this
            page. Continuing to use the service after a change means you accept the updated terms.
          </p>

          <h2 style={H2}>9. Contact</h2>
          <p style={P}>
            Questions about these terms can go through the <Link href="/support">Support page</Link> or the
            repository&apos;s issue tracker. For how data is handled, see the <Link href="/privacy">Privacy Policy</Link>.
          </p>
        </div>
      </main>
    </AppShell>
  )
}
