// Judul dan deskripsi per halaman dokumentasi. Dipakai untuk metadata SEO
// sekaligus sitemap, jadi tidak ada JSX di sini (aman diimpor dari server).
export interface DocMeta {
  title: string
  description: string
}

export const docMeta: Record<string, DocMeta> = {
  "": {
    title: "Documentation",
    description:
      "Everything about only4.fun: GitHub sign-in, deploying a static site or SSR app, ZIP uploads, environment variables, custom domains, the build pipeline, and the HTTP API reference.",
  },
  what: {
    title: "What is only4.fun?",
    description:
      "only4.fun is a small open-source hosting platform that builds your GitHub repository or ZIP archive and serves the result on a free subdomain such as my-project.only4.fun.",
  },
  why: {
    title: "Why it exists",
    description:
      "Why only4.fun exists: free static and SSR hosting with no per-seat pricing, no bandwidth caps, and no cold-start throttling, built as an open-source project.",
  },
  vercel: {
    title: "A free Vercel alternative",
    description:
      "How only4.fun compares with Vercel for small static and server-rendered sites: one subdomain per project, GitHub deploys, custom domains, and no usage meter.",
  },
  opensource: {
    title: "Open source and self-hostable",
    description:
      "only4.fun is open source and can be self-hosted: the Go deployment engine, the Next.js dashboard, SQLite storage, and Docker builds all run on your own server.",
  },
  how: {
    title: "How it works",
    description:
      "The path a deployment takes: GitHub OAuth, repository clone, Docker or local build, release folders on disk, DNS record creation, and serving through the reverse proxy.",
  },
  accounts: {
    title: "Accounts and sign-in",
    description:
      "Sign in with GitHub — no password or registration form. Sessions last 24 hours and the login page remembers which locked page you came from.",
  },
  deploy: {
    title: "Deploying a site",
    description:
      "Deploy from a GitHub repository or upload a ZIP archive, choose the project name that becomes your subdomain, and watch the build finish on the project page.",
  },
  pipeline: {
    title: "Build pipeline",
    description:
      "How only4.fun installs dependencies, runs the framework build, detects static output versus SSR entry points, and publishes each deployment as a release folder.",
  },
  frameworks: {
    title: "Supported frameworks",
    description:
      "Static HTML, Vite, Next.js, Nuxt, SvelteKit, Astro, and other frameworks that only4.fun can build, with the settings each one needs.",
  },
  subdomain: {
    title: "Project name and subdomain",
    description:
      "Project names become subdomains: lower-case letters, numbers, and dashes, 3-63 characters, with reserved and restricted words rejected before a project is created.",
  },
  env: {
    title: "Environment variables",
    description:
      "Environment variables are encrypted at rest, injected at build time, and never sent back to the browser — only key names are shown.",
  },
  database: {
    title: "Database with Moogo",
    description:
      "only4.fun has no built-in database. Pair a static site with Moogo, SQLite over HTTP and open source, to read and write data from a hosted page.",
  },
  status: {
    title: "Deployment statuses",
    description: "What each deployment status means — queued, building, live, or failed — and where to look when a build stops.",
  },
  detail: {
    title: "Project detail page",
    description:
      "The project page shows source, subdomain, custom domain, status, repository metadata, traffic, deployment history, environment variables, and webhook settings.",
  },
  visits: {
    title: "Visit analytics",
    description:
      "Aggregate page views per project per day, with a 14-day chart and per-project totals. No visitor profiles or personal data are collected.",
  },
  manage: {
    title: "Redeploy and delete",
    description:
      "Redeploy the current repository state, roll back to one of the last three releases, clear deployment history, or delete a project and its subdomain.",
  },
  domains: {
    title: "Domains and SSL",
    description:
      "Point your own domain at a project with a CNAME record; SSL is terminated at the proxy or CDN in front of the server, such as Cloudflare.",
  },
  ssr: {
    title: "SSR support",
    description:
      "Server-rendered frameworks run as Node processes behind a reverse proxy, with auto-start, idle timeout, port allocation, and zero-downtime swaps.",
  },
  api: {
    title: "API reference",
    description:
      "Every HTTP endpoint of the platform: authentication, projects, deployments, environment variables, domains, webhooks, and GitHub repository listing.",
  },
  security: {
    title: "Security and data",
    description:
      "How only4.fun protects accounts and projects: hashed session identifiers, AES-256-GCM encrypted tokens, per-project ownership checks, rate limits, and HMAC-verified webhooks.",
  },
  limits: {
    title: "Limits and scope",
    description:
      "Current limits: ZIP upload size, one build at a time, three release rollbacks, 50 concurrent SSR processes, and no built-in database.",
  },
  dos: {
    title: "Do's",
    description: "Recommended practices for deploying on only4.fun: keep the build static, use environment variables, pin dependencies, and test locally first.",
  },
  donts: {
    title: "Don'ts",
    description: "Things to avoid: committing secrets, relying on writable disk, huge repositories, and expecting guaranteed uptime.",
  },
  contributing: {
    title: "Contributing",
    description: "How to run only4.fun locally, the project structure, and how to send fixes or features to the open-source repository.",
  },
  troubleshooting: {
    title: "Troubleshooting",
    description: "Common problems and fixes: failed builds, missing index.html, sign-in loops, unreachable subdomains, and custom domains that do not resolve.",
  },
}
