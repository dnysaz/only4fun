const updated = 'Updated 26 Nov 2026';

export const infoPages = {
  '/docs': {
    label: 'Docs',
    eyebrow: 'Only4fun / Docs',
    title: 'Everything you need to',
    accent: 'ship.',
    lead:
      'Only4fun turns a folder of static files into a live site. Paste a public GitHub repository or drop a ZIP, and we give you a random address on only4.fun — no sign-in required.',
    updated,
    sections: [
      {
        id: 'quickstart',
        heading: 'Quick start',
        body:
          'Open Instant Launch, choose GitHub link or ZIP file, and add your source. As soon as a source exists, the address preview appears. Press Launch Now! and we will publish the site at exactly that address.',
        items: [
          'Public GitHub repositories work without an account.',
          'ZIP files should contain index.html at the root or in a common static output folder.',
          'HTML, CSS, JavaScript, images, fonts, and small data files are supported.',
          'JavaScript frameworks need an account and a normal deploy.',
        ],
      },
      {
        id: 'address',
        heading: 'Your address',
        body:
          'Instant deploys get a random 8-character subdomain, like asj3jkha.only4.fun. The preview shows a dotted placeholder until you add a source, and the final URL always matches what you saw before launching.',
        items: [
          'You cannot pick, edit, or delete an instant address.',
          'Instant projects have no traffic statistics.',
          'Sign in to choose your own subdomain, see views, and keep full control.',
        ],
      },
      {
        id: 'workspace',
        heading: 'With an account',
        body:
          'Signing in unlocks Your projects and Deploy new. Every project gets a detail page with traffic, deploy history, environment variables, a custom domain field, and a delete button.',
        items: [
          'Your projects is one clean vertical list.',
          'Deploy new accepts the same GitHub or ZIP source, with framework support.',
          'Project detail is the only place where a project can be removed.',
        ],
      },
      {
        id: 'limits',
        heading: 'Good to know',
        body:
          'Only4fun is built for fun frontends, not servers. Keep projects small, public, and respectful. Anything that looks like abuse, malware, phishing, or server-side execution will be stopped.',
      },
    ],
    cta: {
      title: 'Ready when you are.',
      body: 'No account, no setup. Just a static site and a link.',
      href: '/instant',
      label: 'Launch Now!',
    },
  },

  '/status': {
    label: 'Status',
    eyebrow: 'Only4fun / Status',
    title: 'Built to stay',
    accent: 'online.',
    lead:
      'A plain-language view of what is running right now. The platform is intentionally small: one deploy pipeline, one dashboard, and public sites served from the edge.',
    updated,
    badge: 'All systems operational',
    status: [
      { name: 'Deploy API', state: 'Operational' },
      { name: 'Dashboard & sign in', state: 'Operational' },
      { name: 'Public sites', state: 'Operational' },
      { name: 'DNS & addresses', state: 'Operational' },
    ],
    sections: [
      {
        id: 'now',
        heading: 'Right now',
        body:
          'No active incidents. Deploys are being accepted, projects are loading, and public addresses are resolving as expected. If you see something different, try a hard refresh before retrying your deploy.',
      },
      {
        id: 'watch',
        heading: 'What we watch',
        body:
          'We keep an eye on the few moving parts that matter. Each one has a simple rule: requests should be fast, deploys should finish, and a live link should open.',
        items: [
          'API response time and error rate.',
          'Deploy queue length and build failures.',
          'DNS propagation for new subdomains.',
          'Edge delivery for published static files.',
        ],
      },
      {
        id: 'incidents',
        heading: 'Incidents',
        body:
          'When something breaks, this page is updated first with what happened, what is affected, and when it is resolved. Routine maintenance is listed here too, even if no downtime is expected.',
      },
      {
        id: 'help',
        heading: 'Something broken?',
        body:
          'If a deploy is stuck, refresh this page and try again. If it still fails, check that the repository is public or that the ZIP actually contains an index.html file. Framework projects need an account.',
      },
    ],
    cta: {
      title: 'Put it to the test.',
      body: 'The quickest way to know the platform is healthy is to ship something small.',
      href: '/instant',
      label: 'Launch Now!',
    },
  },

  '/terms': {
    label: 'Terms',
    eyebrow: 'Only4fun / Terms',
    title: 'The simple',
    accent: 'rules.',
    lead:
      'These terms explain what you can do with Only4fun and what we expect in return. Plain English, no fine print games.',
    updated,
    sections: [
      {
        id: 'acceptance',
        heading: 'Using the service',
        body:
          'By launching a site or signing in, you agree to these terms. If you do not agree, do not deploy. You must be old enough to use the service in your country and you are responsible for what you upload.',
      },
      {
        id: 'content',
        heading: 'Your content',
        body:
          'You own the files you upload and the sites you publish. You give us permission to store, copy, and serve those files only so the platform can do its job. Do not upload material you do not have the right to share.',
      },
      {
        id: 'acceptable-use',
        heading: 'Acceptable use',
        body:
          'Only4fun is for static HTML, CSS, and JavaScript projects. Do not use it for malware, phishing, spam, harassment, illegal content, or anything that tries to run server-side code or abuse the platform.',
        items: [
          'Keep GitHub repositories public when deploying instantly.',
          'Do not try to bypass rate limits or address reservations.',
          'Do not use instant addresses for permanent production systems.',
        ],
      },
      {
        id: 'availability',
        heading: 'Addresses & availability',
        body:
          'Instant addresses are random and temporary by design. They cannot be edited or deleted, and no traffic statistics are provided. Signed-in projects get more control, but we still cannot promise uninterrupted service at all times.',
      },
      {
        id: 'changes',
        heading: 'Changes & removal',
        body:
          'We may update these terms as the service grows. We may also pause or remove a site that breaks the rules or puts the platform at risk. If you have an account, you can delete your own projects from the project detail page.',
      },
      {
        id: 'contact',
        heading: 'Contact',
        body:
          'Questions about these terms can be sent to hello@only4.fun. We read everything, but replies may take a little time because this is a small project.',
      },
    ],
    cta: {
      title: 'Prefer the short version?',
      body: 'See how we handle your data on the Privacy page.',
      href: '/privacy',
      label: 'Read Privacy',
    },
  },

  '/privacy': {
    label: 'Privacy',
    eyebrow: 'Only4fun / Privacy',
    title: 'Your stuff stays',
    accent: 'yours.',
    lead:
      'We collect the minimum needed to run the service. No ad tracking, no selling your data, and no mysterious analytics profile following you around.',
    updated,
    sections: [
      {
        id: 'collect',
        heading: 'What we collect',
        body:
          'When you sign in with GitHub we receive the profile details GitHub shares, such as your username, avatar, and public email if you allow it. When you deploy, we store the source reference or uploaded ZIP, project status, view counts, and settings you add.',
        items: [
          'GitHub profile details needed to identify your account.',
          'Repository URL or ZIP contents needed to build your site.',
          'Project name, subdomain, status, and view counts.',
          'Environment variables and custom domains you choose to save.',
          'Short-lived IP and request logs used for security and rate limiting.',
        ],
      },
      {
        id: 'use',
        heading: 'How we use it',
        body:
          'Your data is used to publish sites, show your dashboard, count views, prevent abuse, and debug failures. We do not use your project files to train models or build advertising profiles.',
      },
      {
        id: 'cookies',
        heading: 'Cookies & sessions',
        body:
          'A session cookie keeps you signed in after GitHub authorization. There are no third-party advertising cookies and no embedded analytics trackers on the dashboard.',
      },
      {
        id: 'sharing',
        heading: 'Who sees it',
        body:
          'We share only what is necessary with GitHub for sign-in and with the infrastructure providers that store, build, and serve the files. Published static sites are public by nature — anything you deploy can be opened by anyone with the link.',
      },
      {
        id: 'choices',
        heading: 'Your choices',
        body:
          'Signed-in users can delete projects from the project detail page. Instant projects have no owner attached, so they cannot be edited or removed. Log files rotate automatically, and you can email us to ask about your account data.',
      },
      {
        id: 'children',
        heading: 'Children',
        body:
          'Only4fun is not aimed at children under 13. If you believe a child has shared personal information here, contact us and we will remove it.',
      },
    ],
    cta: {
      title: 'Want the ground rules?',
      body: 'Terms and Privacy go together. Read the terms before you ship.',
      href: '/terms',
      label: 'Read Terms',
    },
  },

  '/support': {
    label: 'Support',
    eyebrow: 'Only4fun / Support',
    title: 'Need a hand?',
    accent: 'We are here.',
    lead:
      'Stuck on a deploy or wondering why a link is not live? Start with the quick checks below, then send us the details. Small project, real replies.',
    updated,
    sections: [
      {
        id: 'before',
        heading: 'Before you write',
        body:
          'Most issues are solved in under a minute. Work through these checks first so we can focus on the part that is actually broken.',
        items: [
          'Open the Status page and confirm all systems are operational.',
          'Hard refresh the browser and try the link again.',
          'For instant deploy, make sure the GitHub repo is public.',
          'For ZIP uploads, make sure index.html is at the root or in a common static output folder.',
          'Framework projects need an account — instant deploy is static only.',
        ],
      },
      {
        id: 'common',
        heading: 'Common fixes',
        body:
          'A few patterns cover almost every support message we get. Try the matching fix before emailing.',
        items: [
          'Blank page: the files landed one folder too deep; move index.html up and deploy again.',
          'Deploy stuck: refresh Status, then retry with the same source.',
          'Address not opening: new subdomains can take a few minutes to resolve.',
          '404 after publish: check the case of your filenames — Linux is case-sensitive.',
          'No stats on an instant site: that is by design; sign in for traffic numbers.',
        ],
      },
      {
        id: 'write',
        heading: 'What to send us',
        body:
          'If the checks above do not help, email with enough detail to reproduce the problem. The more specific you are, the faster we can look into it.',
        items: [
          'The deployed URL or project name.',
          'The GitHub link, or a note that the source was a ZIP.',
          'What you expected to happen and what happened instead.',
          'Your browser and the rough time of the attempt.',
          'A screenshot if the page looks wrong.',
        ],
      },
      {
        id: 'contact',
        heading: 'Contact',
        body:
          'Write to hello@only4.fun. Please do not include passwords, tokens, or environment variable values — we will never ask for those.',
      },
      {
        id: 'response',
        heading: 'Response time',
        body:
          'Only4fun is a small side project, so replies usually go out within a couple of days. Security reports are treated as urgent and get priority.',
      },
    ],
    cta: {
      title: 'Maybe Docs already has it.',
      body: 'The docs cover deploy, addresses, and account features step by step.',
      href: '/docs',
      label: 'Read Docs',
    },
  },
};

export function getInfoPage(route) {
  return infoPages[route] ?? null;
}
