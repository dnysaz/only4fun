import React from "react"
import Link from "next/link"

const H2: React.CSSProperties = { fontSize: "16px", fontWeight: 650, margin: "34px 0 12px", paddingTop: 24, borderTop: "1px solid #e4e4e4", letterSpacing: "-0.1px" }
const H3: React.CSSProperties = { fontSize: "14px", fontWeight: 650, margin: "24px 0 6px" }
const P: React.CSSProperties = { fontSize: "13.5px", lineHeight: 1.75, color: "#4f4f4f", margin: "0 0 12px" }
const UL: React.CSSProperties = { margin: "0 0 14px", paddingLeft: 22 }
const LI: React.CSSProperties = { fontSize: "13.5px", lineHeight: 1.75, color: "#4f4f4f", marginBottom: 6 }
const CODE: React.CSSProperties = { background: "#f4f4f4", padding: "1px 5px", borderRadius: 4, fontSize: "12.5px", fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace" }
const TABLE: React.CSSProperties = { width: "100%", borderCollapse: "collapse", margin: "0 0 14px", fontSize: "13px" }
const TH: React.CSSProperties = { textAlign: "left", padding: "8px 10px", borderBottom: "1px solid #e4e4e4", fontWeight: 600, color: "#2e2e2e" }
const TD: React.CSSProperties = { padding: "8px 10px", borderBottom: "1px solid #ececec", color: "#4f4f4f", verticalAlign: "top" }

export const docSections: Record<string, { title: string; content: React.ReactNode }> = {
  "": {
    title: "Documentation",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Documentation</h1>
        <p style={{ fontSize: "15px", lineHeight: 1.7, color: "#4f4f4f", margin: "0 0 28px" }}>
          Everything about only4.fun — from the first sign-in to the build log of your tenth deployment.
        </p>
        <p style={P}>
          only4.fun is an open-source, self-hostable hosting platform for static and server-rendered websites. Point it
          at a GitHub repository or drop a ZIP archive, let it build on the server, and the result goes live on a
          subdomain such as <code style={CODE}>my-project.only4.fun</code> — or on a domain you own.
        </p>
        <p style={P}>
          On top of hosting you get: auto-deploy when you push (GitHub webhook), rollback to any of the last three
          releases, environment variables injected at build time, per-project custom domains with SSL, and daily
          visit analytics.
        </p>
        <p style={P}>
          Use the sidebar on the left to navigate to any topic. On a phone, open the <strong>Browse documentation</strong>{" "}
          panel above the article instead, and open the ☰ menu for the dashboard navigation.
        </p>
      </>
    ),
  },
  what: {
    title: "What is only4.fun?",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>What is only4.fun?</h1>
        <p style={P}>
          only4.fun is a small hosting platform for static and server-rendered websites. Point it at a GitHub
          repository or upload a ZIP archive, let it build on the server, and it serves the result on a subdomain such
          as <code style={CODE}>my-project.only4.fun</code>.
        </p>
        <p style={P}>
          A Git repository is the usual source of truth: every deploy starts from a commit that already exists on
          GitHub, and a webhook can start a build automatically whenever you push to the default branch. ZIP projects
          are the exception — they skip the build and serve the archive as-is, so publish a new project when you need
          a new version.
        </p>
        <p style={P}>
          The platform supports both static sites and SSR (Server-Side Rendering) frameworks like Next.js, Nuxt,
          SvelteKit, and more. Static sites are served as files; SSR sites run as Node.js processes behind a
          reverse proxy.
        </p>
        <p style={P}>
          There is no database attached to a project. When a site needs to store data, pair it with{" "}
          <a href="https://moogo.dev" target="_blank" rel="noreferrer" style={{ color: "#2563eb", textDecoration: "none" }}>
            Moogo
          </a>{" "}
          — SQLite over HTTP, free and open source, made by the same author. See{" "}
          <Link href="/docs/database" style={{ color: "#2563eb", textDecoration: "none" }}>Database with Moogo</Link>.
        </p>
      </>
    ),
  },
  why: {
    title: "Why it exists",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Why it exists</h1>
        <p style={P}>
          Large hosting platforms solve every problem, and charge you in concepts: regions, build caches, environment
          scopes, preview branches, pricing tables. For a portfolio, a landing page, a documentation site, or a demo,
          that is a lot of surface for something that is ultimately a folder of files behind a URL.
        </p>
        <p style={P}>only4.fun keeps five promises:</p>
        <ul style={UL}>
          <li style={LI}>Sign in with GitHub. There is no separate account, password, or email confirmation.</li>
          <li style={LI}>Deploy from a repository you already own, public or private — or drop a ZIP archive.</li>
          <li style={LI}>Get a predictable subdomain under a single domain, and connect your own domain later.</li>
          <li style={LI}>Declare environment variables in one plain text box.</li>
          <li style={LI}>Auto-deploy on push with a webhook, and roll back to the previous release when something breaks.</li>
        </ul>
      </>
    ),
  },
  vercel: {
    title: "Vercel alternative",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Vercel alternative</h1>
        <p style={P}>
          only4.fun is designed as an open-source alternative to Vercel. While Vercel is a commercial platform with
          advanced features, only4.fun provides the core deployment experience with full control over your
          infrastructure.
        </p>
        <table style={TABLE}>
          <thead>
            <tr>
              <th style={TH}>Feature</th>
              <th style={TH}>Vercel</th>
              <th style={TH}>only4.fun</th>
            </tr>
          </thead>
          <tbody>
            <tr><td style={TD}>Static site hosting</td><td style={TD}>✅</td><td style={TD}>✅</td></tr>
            <tr><td style={TD}>SSR (Next.js, Nuxt, etc.)</td><td style={TD}>✅</td><td style={TD}>✅ (via Node.js process)</td></tr>
            <tr><td style={TD}>Auto-deploy on git push</td><td style={TD}>✅</td><td style={TD}>✅ (GitHub webhook)</td></tr>
            <tr><td style={TD}>Deploy without Git (ZIP upload)</td><td style={TD}>✅</td><td style={TD}>✅</td></tr>
            <tr><td style={TD}>Rollback to a previous build</td><td style={TD}>✅</td><td style={TD}>✅ (3 releases kept)</td></tr>
            <tr><td style={TD}>Preview deployments</td><td style={TD}>✅</td><td style={TD}>❌</td></tr>
            <tr><td style={TD}>Custom domains</td><td style={TD}>✅</td><td style={TD}>✅ (CNAME + Cloudflare SSL)</td></tr>
            <tr><td style={TD}>Edge functions</td><td style={TD}>✅</td><td style={TD}>❌</td></tr>
            <tr><td style={TD}>Image optimization</td><td style={TD}>✅</td><td style={TD}>❌</td></tr>
            <tr><td style={TD}>Self-hostable</td><td style={TD}>❌</td><td style={TD}>✅</td></tr>
            <tr><td style={TD}>Open source</td><td style={TD}>❌</td><td style={TD}>✅</td></tr>
            <tr><td style={TD}>Free (no limits)</td><td style={TD}>❌ (paid tiers)</td><td style={TD}>✅</td></tr>
            <tr><td style={TD}>Data ownership</td><td style={TD}>❌ (cloud)</td><td style={TD}>✅ (your server)</td></tr>
          </tbody>
        </table>
      </>
    ),
  },
  opensource: {
    title: "Open source",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Open source</h1>
        <p style={P}>
          only4.fun is fully open source. You can deploy it on your own VPS, customize it, and contribute back to
          the project. The source code is available on GitHub.
        </p>
        <p style={P}>
          <strong>Self-deployment:</strong> Follow the deployment guide in <code style={CODE}>deploy.md</code> to
          set up on your own VPS with Cloudflare for SSL and proxy.
        </p>
        <p style={P}>
          <strong>No vendor lock-in:</strong> Your data stays on your server. The database is a single SQLite file,
          and your site files are in a regular directory. You can back up, migrate, or delete everything at any time.
        </p>
      </>
    ),
  },
  how: {
    title: "How it works",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>How a deployment works</h1>
        <p style={P}>Four parts are involved, and each one has a single job:</p>
        <table style={TABLE}>
          <thead>
            <tr>
              <th style={TH}>Part</th>
              <th style={TH}>Responsibility</th>
            </tr>
          </thead>
          <tbody>
            <tr><td style={TD}>Dashboard</td><td style={TD}>Sign in, pick a repository or upload a ZIP, watch status, read logs, edit variables</td></tr>
            <tr><td style={TD}>API server</td><td style={TD}>GitHub OAuth, sessions, projects, environment variables, deployment records</td></tr>
            <tr><td style={TD}>Build worker</td><td style={TD}>Clone the repository, install dependencies, run the build, publish the output</td></tr>
            <tr><td style={TD}>Edge and DNS</td><td style={TD}>Serve the subdomain to visitors, terminate TLS (via Cloudflare)</td></tr>
          </tbody>
        </table>
        <p style={P}>
          When a build succeeds, the files are published atomically: every build becomes its own release folder and a
          single symlink swap makes it live, so visitors never see a half-written site. The three most recent releases
          stay on disk, which is what powers rollback.
        </p>
      </>
    ),
  },
  accounts: {
    title: "Accounts and sign-in",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Accounts and sign-in</h1>
        <p style={P}>
          There is no registration form. Your account is created the first time you authorize the app with GitHub.
        </p>
        <p style={P}>The authorization asks for two scopes:</p>
        <ul style={UL}>
          <li style={LI}><code style={CODE}>read:user</code> and <code style={CODE}>user:email</code> to identify you.</li>
          <li style={LI}>Repository access to list and clone your repositories, including private ones.</li>
        </ul>
        <p style={P}>
          Sign-in always starts at the <code style={CODE}>/login</code> page, which has a single button:{" "}
          <strong>Continue with GitHub</strong>. The home page and the public bar (Docs, Help, Support) link there too.
          If you already have a session, <code style={CODE}>/login</code> sends you straight to the dashboard.
        </p>
        <p style={P}>
          The session lasts 24 hours. The browser receives an opaque session identifier in an{" "}
          <code style={CODE}>HttpOnly</code> cookie. On the server only a SHA-256 hash is stored, and your GitHub
          access token is encrypted at rest with AES-256-GCM.
        </p>
        <p style={P}>
          Only the home page and the Docs, Help, and Support pages are public. Dashboard, New project, project detail,
          and Profile need a session: opening them while signed out shows a short "checking your session" screen and
          then sends you to <code style={CODE}>/login</code>. The dashboard navigation is not rendered at all in that
          state, so nothing flashes before the redirect.
        </p>
        <p style={P}>
          The login page remembers where you were going. Locked pages link to <code style={CODE}>/login?next=/path</code>,
          the API stores that path in a short-lived <code style={CODE}>oauth_next</code> cookie during the OAuth round
          trip, and the callback returns you to it instead of the dashboard. Only local paths are accepted, so the
          parameter cannot be used to redirect you to another site. The cookie is deleted as soon as it is read.
        </p>
        <p style={P}>
          The <strong>Profile</strong> page in the sidebar shows your GitHub username, email, and sign-in details, and
          is where you sign out. Signing out deletes the session on the server, clears the cookie in this browser, and
          returns you to <code style={CODE}>/login</code>.
        </p>
      </>
    ),
  },
  deploy: {
    title: "Deploying a site",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Deploying a site</h1>
        <ol style={UL}>
          <li style={LI}>In the sidebar (on a phone, open the ☰ menu first), choose New project.</li>
          <li style={LI}><strong>Source</strong> — pick a GitHub repository (private ones are marked) or switch to <strong>Upload ZIP</strong> and drop a static site archive instead.</li>
          <li style={LI}><strong>Project name</strong> — becomes the subdomain. Typing <code style={CODE}>portfolio</code> gives <code style={CODE}>portfolio.only4.fun</code>.</li>
          <li style={LI}><strong>Environment variables</strong> — optional, added later from the project page.</li>
          <li style={LI}><strong>Review</strong> — the last block repeats the selected source and the final URL before anything runs.</li>
          <li style={LI}>Click Deploy project. You are taken to the project page while the build runs.</li>
        </ol>
        <p style={P}>
          The build log is the first place to look when something goes wrong. ZIP uploads skip the build step and must
          contain <code style={CODE}>index.html</code> at the root (or inside a single top-level folder).
        </p>
        <p style={P}>
          You can also connect a GitHub webhook on the project page so every push to the default branch deploys
          automatically.
        </p>
      </>
    ),
  },
  pipeline: {
    title: "Build pipeline",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Build pipeline</h1>
        <p style={P}>Each job runs the same sequence:</p>
        <ol style={UL}>
          <li style={LI}><strong>Clone</strong> — shallow clone of the default branch.</li>
          <li style={LI}><strong>Detect</strong> — if no <code style={CODE}>package.json</code> and there is an <code style={CODE}>index.html</code>, served as-is.</li>
          <li style={LI}><strong>Install</strong> — lockfile decides the tool: npm ci, pnpm install, yarn install, or npm install.</li>
          <li style={LI}><strong>Build</strong> — runs <code style={CODE}>npm run build</code> when declared.</li>
          <li style={LI}><strong>Collect output</strong> — first directory with <code style={CODE}>index.html</code>: dist/, out/, build/, public/, .next/, then root.</li>
          <li style={LI}><strong>Publish</strong> — copy into a release folder, skip node_modules and .git, swap the symlink atomically (three releases kept).</li>
          <li style={LI}><strong>DNS</strong> — create/update subdomain record when Cloudflare is configured.</li>
        </ol>
        <p style={P}>
          Builds run in Docker by default. When Docker is unavailable the build fails unless the operator explicitly
          sets <code style={CODE}>BUILDER_ALLOW_LOCAL=true</code>, because repository scripts would otherwise run
          directly on the server. A build exceeding the timeout is stopped and marked failed. Jobs are queued and
          processed one at a time.
        </p>
      </>
    ),
  },
  frameworks: {
    title: "Supported frameworks",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Supported frameworks</h1>
        <p style={P}>The build pipeline auto-detects your framework and uses the appropriate build command:</p>
        <table style={TABLE}>
          <thead>
            <tr>
              <th style={TH}>Framework</th>
              <th style={TH}>Detection</th>
              <th style={TH}>Output dir</th>
            </tr>
          </thead>
          <tbody>
            <tr><td style={TD}>Next.js</td><td style={TD}><code style={CODE}>next</code> in deps</td><td style={TD}><code style={CODE}>out/</code> or <code style={CODE}>.next/</code></td></tr>
            <tr><td style={TD}>Nuxt</td><td style={TD}><code style={CODE}>nuxt</code> in deps</td><td style={TD}><code style={CODE}>dist/</code> or <code style={CODE}>.output/</code></td></tr>
            <tr><td style={TD}>Astro</td><td style={TD}><code style={CODE}>astro</code> in deps</td><td style={TD}><code style={CODE}>dist/</code></td></tr>
            <tr><td style={TD}>SvelteKit</td><td style={TD}><code style={CODE}>@sveltejs/kit</code></td><td style={TD}><code style={CODE}>build/</code></td></tr>
            <tr><td style={TD}>Remix</td><td style={TD}><code style={CODE}>@remix-run</code></td><td style={TD}><code style={CODE}>build/</code></td></tr>
            <tr><td style={TD}>Vite (Vue/React)</td><td style={TD}><code style={CODE}>vite</code> in deps</td><td style={TD}><code style={CODE}>dist/</code></td></tr>
            <tr><td style={TD}>React (CRA)</td><td style={TD}><code style={CODE}>react</code> in deps</td><td style={TD}><code style={CODE}>build/</code></td></tr>
            <tr><td style={TD}>Angular</td><td style={TD}><code style={CODE}>@angular/core</code></td><td style={TD}><code style={CODE}>dist/</code></td></tr>
            <tr><td style={TD}>Gatsby</td><td style={TD}><code style={CODE}>gatsby</code></td><td style={TD}><code style={CODE}>public/</code></td></tr>
            <tr><td style={TD}>Docusaurus</td><td style={TD}><code style={CODE}>@docusaurus/core</code></td><td style={TD}><code style={CODE}>build/</code></td></tr>
            <tr><td style={TD}>Static (no package.json)</td><td style={TD}>—</td><td style={TD}>Repository root</td></tr>
          </tbody>
        </table>
        <p style={P}>
          <strong>Package managers:</strong> npm, yarn, pnpm, bun — auto-detected from lockfile.
        </p>
        <p style={P}>
          <strong>Build-time HTTP requests:</strong> Environment variables are injected during build. Frameworks can
          fetch data from Supabase, Turso, or any API during the build process.
        </p>
        <p style={P}>
          <strong>SSR support:</strong> For SSR frameworks, the build must expose a start script in
          <code style={CODE}>package.json</code> and a server entry point (for example
          <code style={CODE}>server.js</code> or <code style={CODE}>.next/standalone/server.js</code>). The platform will
          run it as a Node.js process and proxy requests to it. For static-only deployment, enable static export in
          your framework config.
        </p>
      </>
    ),
  },
  subdomain: {
    title: "Project name and subdomain",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Project name and subdomain</h1>
        <p style={P}>
          The project name is the only part of the URL you choose, and it must be unique across the platform.
        </p>
        <ul style={UL}>
          <li style={LI}>Allowed characters: <code style={CODE}>a-z</code>, <code style={CODE}>0-9</code>, and dash.</li>
          <li style={LI}>Must be at least 3 characters, must not start or end with a dash, max 63 characters.</li>
          <li style={LI}>Forbidden words: profanity (EN/ID), administrative terms, pornographic terms, political terms, public figure names, platform-reserved names.</li>
          <li style={LI}>Reserved: <code style={CODE}>www</code>, <code style={CODE}>api</code>, <code style={CODE}>app</code>, <code style={CODE}>admin</code>, <code style={CODE}>auth</code>, <code style={CODE}>docs</code>, <code style={CODE}>help</code>, <code style={CODE}>profile</code>, <code style={CODE}>dashboard</code>, <code style={CODE}>mail</code>.</li>
        </ul>
        <p style={P}>
          If a name is taken, the API answers with a conflict. Deleting a project frees its name.
        </p>
      </>
    ),
  },
  env: {
    title: "Environment variables",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Environment variables</h1>
        <p style={P}>
          Variables are for values the build needs but that should not live in the repository. They are stored with the
          project and injected into the build process.
        </p>
        <p style={P}>
          Values are build-time only. Anything compiled into a frontend bundle is readable by visitors, so never put a
          long-lived secret in a static site. Changing a value has no effect on an already built site; redeploy to rebuild.
        </p>
        <p style={P}>
          Values are encrypted at rest (AES-256-GCM) and never sent to the browser. Only keys are displayed. When
          editing, entering a key with a new value replaces the old value. Keys not listed are kept. Empty value deletes
          the key.
        </p>
        <p style={P}>
          The API never returns values. Only the owner can read keys, and editing happens in a side panel opened from
          the project page (<strong>Environment variables → Configure</strong>).
        </p>
      </>
    ),
  },
  database: {
    title: "Database with Moogo",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Database with Moogo</h1>
        <p style={P}>
          only4.fun serves static files and, optionally, a Node process for SSR projects. It does not include a
          database: there is no MySQL, Postgres, or Redis attached to your project, and nothing written during a build
          survives the next deployment.
        </p>
        <p style={P}>
          For projects that need one, the recommended companion is{" "}
          <a href="https://moogo.dev" target="_blank" rel="noreferrer" style={{ color: "#2563eb", textDecoration: "none" }}>
            <strong>Moogo</strong>
          </a>{" "}
          — <strong>SQLite over HTTP</strong>, also <strong>free and open source</strong> and made by the same author as
          only4.fun. Every Moogo project gets its own SQLite file and its own key, plus 256 MB of object storage, and
          you reach all of it with plain HTTP requests: no connection string, no driver, no connection pool.
        </p>

        <h2 style={H2}>Why it fits a static site</h2>
        <ul style={UL}>
          <li style={LI}>
            A page can read and write with <code style={CODE}>fetch()</code>; there is no server to run and nothing
            extra to deploy next to your site.
          </li>
          <li style={LI}>The database lives outside the build, so data survives redeploys.</li>
          <li style={LI}>
            It is ordinary SQLite exposed as SQL over HTTP, with a dashboard, scheduled backups, and an object storage
            bucket per project.
          </li>
          <li style={LI}>
            Like only4.fun, it is an open-source project built in public — free to use, without an uptime guarantee.
          </li>
        </ul>

        <h2 style={H2}>Credentials</h2>
        <p style={P}>Three values identify your database:</p>
        <ul style={UL}>
          <li style={LI}><code style={CODE}>MOOGO_BASE_URL</code> — the API origin, for example <code style={CODE}>https://api.moogo.dev</code>.</li>
          <li style={LI}><code style={CODE}>MOOGO_PROJECT_ID</code> — your project id from the dashboard.</li>
          <li style={LI}><code style={CODE}>MOOGO_SQL_KEY</code> — the project key, sent as <code style={CODE}>Authorization: Bearer moogo_...</code>.</li>
        </ul>
        <p style={P}>
          The project URL is <code style={CODE}>$MOOGO_BASE_URL/p/$MOOGO_PROJECT_ID</code>. Add the three values to the{" "}
          <strong>Environment variables</strong> panel of your project, then redeploy so the build can read them.
        </p>

        <h2 style={H2}>Three endpoints</h2>
        <table style={TABLE}>
          <thead>
            <tr>
              <th style={TH}>Endpoint</th>
              <th style={TH}>Use it for</th>
            </tr>
          </thead>
          <tbody>
            <tr><td style={TD}><code style={CODE}>POST /p/:id/query</code></td><td style={TD}>Reads — <code style={CODE}>SELECT</code>, <code style={CODE}>VALUES</code>, <code style={CODE}>PRAGMA</code>, <code style={CODE}>EXPLAIN</code></td></tr>
            <tr><td style={TD}><code style={CODE}>POST /p/:id/exec</code></td><td style={TD}>Writes and DDL — <code style={CODE}>INSERT</code>, <code style={CODE}>UPDATE</code>, <code style={CODE}>DELETE</code>, <code style={CODE}>CREATE</code>, <code style={CODE}>ALTER</code>, <code style={CODE}>DROP</code></td></tr>
            <tr><td style={TD}><code style={CODE}>POST /p/:id/transaction</code></td><td style={TD}>Several writes committed together, up to 100 statements</td></tr>
          </tbody>
        </table>
        <p style={P}>
          Both read and write endpoints take <code style={CODE}>{`{ "query": "...", "args": [...] }`}</code>. Always put
          values in <code style={CODE}>args</code> — they travel as bound parameters, so a value can never become SQL
          syntax.
        </p>

        <h2 style={H2}>Reading from a page</h2>
        <pre style={{ background: "#f7f7f7", border: "1px solid #ececec", borderRadius: 8, padding: "12px 14px", overflowX: "auto", fontSize: "12.5px", lineHeight: 1.6, color: "#3d3d3d", margin: "0 0 14px" }}>
          <code>{`const base = \`\${MOOGO_BASE_URL}/p/\${MOOGO_PROJECT_ID}\`

const res = await fetch(\`\${base}/query\`, {
  method: "POST",
  headers: {
    Authorization: \`Bearer \${MOOGO_SQL_KEY}\`,
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    query: "select id, title from posts order by id desc limit 20",
  }),
})

const { columns, rows } = await res.json()
// rows are arrays aligned with columns: rows[0][0] belongs to columns[0]`}</code>
        </pre>
        <p style={P}>
          A read answers with <code style={CODE}>columns</code> and <code style={CODE}>rows</code> (arrays, not
          objects), plus <code style={CODE}>row_count</code> and <code style={CODE}>truncated</code> — check the flag,
          because reads stop at 1000 rows or 16 MB. A write answers with{" "}
          <code style={CODE}>rows_affected</code> and <code style={CODE}>last_insert_rowid</code>, and{" "}
          <code style={CODE}>RETURNING</code> can hand back the rows you just changed in the same call.
        </p>

        <h2 style={H2}>Keep it safe</h2>
        <ul style={UL}>
          <li style={LI}>
            Anything compiled into a frontend bundle is public. A key used in client-side code can be read by any
            visitor, so only do that for data you intend to be publicly readable.
          </li>
          <li style={LI}>
            For writes or private data, route the requests through an SSR project or another small server you control,
            and keep the key there.
          </li>
          <li style={LI}>
            Moogo, like only4.fun, is a free open-source project: no warranty and no uptime guarantee. Keep your own
            backups of anything important.
          </li>
        </ul>
        <p style={P}>
          Full reference — SQL API, object storage, JavaScript client, limits, and error codes — is at{" "}
          <a href="https://moogo.dev" target="_blank" rel="noreferrer" style={{ color: "#2563eb", textDecoration: "none" }}>
            moogo.dev
          </a>.
        </p>
      </>
    ),
  },
  status: {
    title: "Deployment statuses",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Deployment statuses</h1>
        <table style={TABLE}>
          <thead>
            <tr>
              <th style={TH}>Status</th>
              <th style={TH}>Shown as</th>
              <th style={TH}>Meaning</th>
            </tr>
          </thead>
          <tbody>
            <tr><td style={TD}>pending</td><td style={TD}>Queued</td><td style={TD}>No build has started yet.</td></tr>
            <tr><td style={TD}>building</td><td style={TD}>Building</td><td style={TD}>The repository is being cloned, installed, or built.</td></tr>
            <tr><td style={TD}>deployed</td><td style={TD}>Live</td><td style={TD}>The build succeeded and the subdomain serves the output.</td></tr>
            <tr><td style={TD}>failed</td><td style={TD}>Error</td><td style={TD}>The build or the DNS step failed. The log explains why.</td></tr>
          </tbody>
        </table>
        <p style={P}>
          The same labels are used on the dashboard, on the project page, and next to each entry in the deployment
          history. While a project is queued or building, the project page refreshes itself.
        </p>
      </>
    ),
  },
  detail: {
    title: "Project detail page",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>The project detail page</h1>
        <ul style={UL}>
          <li style={LI}><strong>Header</strong> — subdomain, status, the live URL, and the Redeploy, Clear history, and Delete actions (delete asks for confirmation in a modal).</li>
          <li style={LI}><strong>Overview</strong> — repository (or ZIP source), subdomain, custom domain, status, creation and update timestamps, plus repository metadata from GitHub: description, visibility, default branch, language, stars, forks, open issues, last push.</li>
          <li style={LI}><strong>Traffic</strong> — total views, last 14 days, daily average, peak day, the daily chart, and the last visit.</li>
          <li style={LI}><strong>Deployments</strong> — every run with status, timestamp, and full log. Last 20 runs kept; each successful run offers <strong>Rollback</strong> and <strong>View log</strong>.</li>
          <li style={LI}><strong>Environment variables</strong>, <strong>Custom domain</strong>, and <strong>Auto-deploy webhook</strong> — one compact row each with a Configure button that opens a side panel sliding in from the right.</li>
          <li style={LI}><strong>Build terminal</strong> — live log viewer shown only while a build is queued or running.</li>
        </ul>
      </>
    ),
  },
  visits: {
    title: "Visit analytics",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Visit analytics</h1>
        <p style={P}>
          Every request that resolves to an HTML page is counted. Stylesheets, scripts, images, and fonts are not
          counted. Totals stored per project, per-day counter for last 14 days. No cookies, no fingerprinting, no
          third-party analytics.
        </p>
      </>
    ),
  },
  manage: {
    title: "Redeploy and delete",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Redeploy and delete</h1>
        <p style={P}><strong>Redeploy</strong> rebuilds from the current repository state using current env vars. Projects created from a ZIP have no repository, so they cannot be redeployed.</p>
        <p style={P}><strong>Rollback</strong> re-activates a previous successful release from the Deployments list. The three most recent releases are kept on disk.</p>
        <p style={P}><strong>Delete</strong> opens a confirmation modal, then removes the project record, published files, releases, and DNS record. The GitHub repository is untouched. Cannot be undone. Subdomain becomes available to others.</p>
        <p style={P}><strong>Clear history</strong> only deletes deployment logs and does not touch the live site.</p>
      </>
    ),
  },
  domains: {
    title: "Domains and SSL",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Domains and SSL</h1>
        <p style={P}>
          Every project gets <code style={CODE}>myproject.only4.fun</code> automatically. You can also connect a domain
          you own: open <strong>Custom domain → Configure</strong> on the project page, enter the hostname, then create
          a <strong>CNAME</strong> record for it that points to <code style={CODE}>myproject.only4.fun</code> (or an A
          record to the server IP). The panel shows the exact target and warns you if the certificate step is not
          active yet.
        </p>
        <p style={P}>
          TLS for platform subdomains is terminated at the edge via Cloudflare. The wildcard record
          <code style={CODE}>*.only4.fun</code> is proxied through Cloudflare, which serves a certificate covering the
          whole wildcard. For custom domains, the server registers the hostname with Cloudflare for SaaS so a
          certificate is issued automatically; if that feature is not enabled on the zone, terminate TLS in your own
          Cloudflare zone or with your own certificate instead.
        </p>
      </>
    ),
  },
  ssr: {
    title: "SSR support",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>SSR support</h1>
        <p style={P}>
          The platform supports Server-Side Rendering via Node.js processes. SSR is only enabled when the repository
          explicitly declares it: <code style={CODE}>package.json</code> must contain a
          <code style={CODE}>start</code> script, and the published output must contain
          <code style={CODE}>server.js</code>, <code style={CODE}>server.mjs</code>,
          <code style={CODE}>build/server.js</code>, <code style={CODE}>dist/server.js</code>, or
          <code style={CODE}>.next/standalone/server.js</code>. When those exist, the platform starts a Node.js process
          and proxies requests to it.
        </p>
        <p style={P}>
          For static-only deployment, enable static export in your framework config (e.g., Next.js:
          <code style={CODE}>output: 'export'</code> in next.config.js).
        </p>
        <p style={P}>
          Process management includes: auto-start on first request, idle timeout (10 min), max 50 concurrent
          processes, port allocation (4000-5000). On redeploy the platform starts a new process on another port and
          only switches traffic once it answers, then stops the old one — visitors keep the old version until the new
          one is ready.
        </p>
      </>
    ),
  },
  api: {
    title: "API reference",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>API reference</h1>
        <table style={TABLE}>
          <thead>
            <tr>
              <th style={TH}>Endpoint</th>
              <th style={TH}>Purpose</th>
            </tr>
          </thead>
          <tbody>
            <tr><td style={TD}><code style={CODE}>GET /health</code></td><td style={TD}>Liveness check, no authentication</td></tr>
            <tr><td style={TD}><code style={CODE}>GET /login</code></td><td style={TD}>Frontend sign-in page with a single Continue with GitHub button (accepts <code style={CODE}>?next=</code>)</td></tr>
            <tr><td style={TD}><code style={CODE}>GET /auth/github/login?next=/path</code></td><td style={TD}>Start the OAuth flow; the local path is kept in the <code style={CODE}>oauth_next</code> cookie</td></tr>
            <tr><td style={TD}><code style={CODE}>GET /auth/github/callback</code></td><td style={TD}>Finish the flow, set the session cookie, and redirect to the remembered path</td></tr>
            <tr><td style={TD}><code style={CODE}>GET /api/auth/me</code></td><td style={TD}>Current user</td></tr>
            <tr><td style={TD}><code style={CODE}>POST /api/auth/logout</code></td><td style={TD}>Delete the session</td></tr>
            <tr><td style={TD}><code style={CODE}>GET /api/projects</code></td><td style={TD}>List your projects</td></tr>
            <tr><td style={TD}><code style={CODE}>POST /api/projects</code></td><td style={TD}>Create a project and queue its build</td></tr>
            <tr><td style={TD}><code style={CODE}>POST /api/projects/upload</code></td><td style={TD}>Create a static project from a ZIP (multipart: subdomain, file)</td></tr>
            <tr><td style={TD}><code style={CODE}>GET /api/projects/:id</code></td><td style={TD}>Project detail with visits</td></tr>
            <tr><td style={TD}><code style={CODE}>DELETE /api/projects/:id</code></td><td style={TD}>Delete project, files, and DNS record</td></tr>
            <tr><td style={TD}><code style={CODE}>POST /api/projects/:id/deploy</code></td><td style={TD}>Queue a rebuild</td></tr>
            <tr><td style={TD}><code style={CODE}>GET /api/projects/:id/deployments</code></td><td style={TD}>Deployment history with logs</td></tr>
            <tr><td style={TD}><code style={CODE}>POST /api/projects/:id/deployments/:deploymentID/rollback</code></td><td style={TD}>Re-activate a previous release</td></tr>
            <tr><td style={TD}><code style={CODE}>DELETE /api/projects/:id/deployments</code></td><td style={TD}>Clear deployment history</td></tr>
            <tr><td style={TD}><code style={CODE}>GET /api/projects/:id/webhook</code></td><td style={TD}>Webhook URL and secret (owner only)</td></tr>
            <tr><td style={TD}><code style={CODE}>PUT /api/projects/:id/domain</code></td><td style={TD}>Set or clear the custom domain</td></tr>
            <tr><td style={TD}><code style={CODE}>POST /webhooks/github</code></td><td style={TD}>GitHub push webhook, verified by HMAC</td></tr>
            <tr><td style={TD}><code style={CODE}>GET /api/projects/:id/repo</code></td><td style={TD}>Repository metadata from GitHub</td></tr>
            <tr><td style={TD}><code style={CODE}>GET /api/projects/:id/env</code></td><td style={TD}>Read env var keys (owner only)</td></tr>
            <tr><td style={TD}><code style={CODE}>PUT /api/projects/:id/env</code></td><td style={TD}>Update env vars (merge, not replace)</td></tr>
            <tr><td style={TD}><code style={CODE}>GET /api/check-subdomain?name=xxx</code></td><td style={TD}>Check subdomain availability</td></tr>
            <tr><td style={TD}><code style={CODE}>GET /api/github/repos?page=N</code></td><td style={TD}>Repositories available to deploy, 100 per page (returns has_more)</td></tr>
          </tbody>
        </table>
      </>
    ),
  },
  security: {
    title: "Security and data",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Security and data</h1>
        <ul style={UL}>
          <li style={LI}>Sessions are random 256-bit identifiers; only a SHA-256 hash is stored.</li>
          <li style={LI}>GitHub access tokens are encrypted with AES-256-GCM.</li>
          <li style={LI}>OAuth callbacks protected by random state in HttpOnly cookie.</li>
          <li style={LI}>Every project endpoint verifies ownership (404 for others).</li>
          <li style={LI}>Creating, deploying, editing, and deleting are rate limited per user.</li>
          <li style={LI}>Environment values encrypted at rest, never sent to browser.</li>
          <li style={LI}>Project names validated against XSS, SQL injection, CSRF, path traversal.</li>
          <li style={LI}>Profanity filter (EN + ID), administrative/porn/political word filter.</li>
          <li style={LI}>GitHub webhooks are verified with an HMAC SHA-256 signature against a per-project secret; the endpoint takes no session.</li>
          <li style={LI}>ZIP uploads are limited (100 MB compressed, 500 MB extracted, 20,000 files) and reject path traversal and symlinks.</li>
          <li style={LI}>Builds run inside Docker by default; repository scripts only run on the host when the operator opts in explicitly.</li>
          <li style={LI}>SQLite database holds accounts, projects, env vars (encrypted), deployment logs. Back it up.</li>
        </ul>
      </>
    ),
  },
  limits: {
    title: "Limits and scope",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Limits and scope</h1>
        <ul style={UL}>
          <li style={LI}>
            Static output only by default. SSR requires a <code style={CODE}>start</code> script in
            <code style={CODE}>package.json</code> plus a server entry point.
          </li>
          <li style={LI}>Custom domains need a DNS record you control; automatic certificates require Cloudflare for SaaS.</li>
          <li style={LI}>Auto-deploy runs from a GitHub webhook you configure per project; other Git hosts are manual.</li>
          <li style={LI}>Builds run one at a time by default, with a configurable timeout.</li>
          <li style={LI}>Repository selection loads 100 repositories per page with a “Load more” button.</li>
          <li style={LI}>Rollback is limited to the three most recent releases on disk; preview deployments are not available.</li>
          <li style={LI}>ZIP uploads: 100 MB compressed, 500 MB after extraction, 20,000 entries. ZIP projects have no repository, so publish a new project instead of redeploying.</li>
          <li style={LI}>Webhook and custom-domain settings are limited to one project each (one webhook per project, one project per domain).</li>
          <li style={LI}>Max 50 concurrent SSR processes.</li>
          <li style={LI}>
            No database is included. For SQLite over HTTP, pair the project with{" "}
            <a href="https://moogo.dev" target="_blank" rel="noreferrer" style={{ color: "#2563eb", textDecoration: "none" }}>
              Moogo
            </a>{" "}
            (free and open source, made by the same author).
          </li>
        </ul>
      </>
    ),
  },
  dos: {
    title: "Do's",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Do's</h1>
        <ul style={UL}>
          <li style={LI}>Use environment variables for API keys, database URLs, and configuration.</li>
          <li style={LI}>Keep your repository's default branch up to date.</li>
          <li style={LI}>Use lockfiles (package-lock.json, yarn.lock, pnpm-lock.yaml) for reproducible builds.</li>
          <li style={LI}>Enable static export for SSR frameworks when possible (better performance).</li>
          <li style={LI}>Redeploy after changing environment variables.</li>
          <li style={LI}>Use descriptive project names (at least 3 characters).</li>
          <li style={LI}>Back up your database regularly.</li>
          <li style={LI}>Set SESSION_SECRET in production.</li>
        </ul>
      </>
    ),
  },
  donts: {
    title: "Don'ts",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Don'ts</h1>
        <ul style={UL}>
          <li style={LI}>Don't put long-lived secrets in frontend bundles (they're readable by visitors).</li>
          <li style={LI}>Don't use reserved names (admin, api, www, etc.) as project names.</li>
          <li style={LI}>Don't use profanity or offensive words in project names.</li>
          <li style={LI}>Don't expose the database file publicly.</li>
          <li style={LI}>Don't change SESSION_SECRET after data is stored (breaks decryption).</li>
          <li style={LI}>Don't run multiple backend instances on the same database.</li>
          <li style={LI}>Don't use the default SESSION_SECRET in production.</li>
          <li style={LI}>Don't set <code style={CODE}>BUILDER_ALLOW_LOCAL=true</code> on a public instance: repository scripts would run on the server.</li>
        </ul>
      </>
    ),
  },
  contributing: {
    title: "Contributing",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Contributing</h1>
        <p style={P}>
          only4.fun is open source. Contributions are welcome! Here's how to get started:
        </p>
        <ol style={UL}>
          <li style={LI}><strong>Fork</strong> the repository on GitHub.</li>
          <li style={LI}><strong>Clone</strong> your fork locally.</li>
          <li style={LI}><strong>Create a branch</strong> for your feature or fix.</li>
          <li style={LI}><strong>Make changes</strong> and test locally.</li>
          <li style={LI}><strong>Run checks</strong>: <code style={CODE}>npx tsc --noEmit</code> (frontend) and <code style={CODE}>go build ./... && go vet ./... && go test ./...</code> (backend). CI runs both on every pull request.</li>
          <li style={LI}><strong>Submit a pull request</strong> with a clear description.</li>
        </ol>
        <p style={P}><strong>Guidelines:</strong></p>
        <ul style={UL}>
          <li style={LI}>Keep the UI simple and consistent.</li>
          <li style={LI}>All user-facing text should be in English.</li>
          <li style={LI}>Validate all inputs on both frontend and backend.</li>
          <li style={LI}>Encrypt sensitive data at rest.</li>
          <li style={LI}>Don't expose secrets to the browser.</li>
          <li style={LI}>Test your changes before submitting.</li>
        </ul>
      </>
    ),
  },
  troubleshooting: {
    title: "Troubleshooting",
    content: (
      <>
        <h1 style={{ fontSize: "27px", fontWeight: 700, margin: "0 0 8px", letterSpacing: "-0.8px" }}>Troubleshooting</h1>
        <h3 style={H3}>The dashboard asks me to sign in again</h3>
        <p style={P}>Sessions expire after 24 hours. Sign in again; if it happens immediately, check that the API host answers on <code style={CODE}>/health</code>.</p>
        <h3 style={H3}>My repository is missing from the list</h3>
        <p style={P}>Repositories load 100 per page: use <strong>Load more repositories</strong>, clear the search box, push a commit, or re-authorize the app.</p>
        <h3 style={H3}>Deploy fails immediately</h3>
        <p style={P}>Open the deployment log. A duplicate project name, invalid repository URL, or unreadable repository are usual causes.</p>
        <h3 style={H3}>The log says local builds are disabled</h3>
        <p style={P}>Docker was not reachable when the job ran, so the server refused to run repository scripts unisolated. Start Docker, or — only on an instance you fully trust — set <code style={CODE}>BUILDER_ALLOW_LOCAL=true</code>.</p>
        <h3 style={H3}>Build fails while installing dependencies</h3>
        <p style={P}>Check that a lockfile matches your manifest. The log shows the failing command and its last lines.</p>
        <h3 style={H3}>The site shows a 404 or the wrong page</h3>
        <p style={P}>The build must produce an <code style={CODE}>index.html</code> in one of the detected output directories.</p>
        <h3 style={H3}>The subdomain does not resolve</h3>
        <p style={P}>Wait for DNS propagation, confirm project status is deployed, and check that the record points at the server.</p>
        <h3 style={H3}>SSR site not working</h3>
        <p style={P}>
          Confirm <code style={CODE}>package.json</code> has a <code style={CODE}>start</code> script and that the build
          publishes a supported server entry point, then redeploy. A repository that only contains an
          <code style={CODE}>index.js</code> file is treated as a static site on purpose.
        </p>
      </>
    ),
  },
}
