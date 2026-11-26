// Route-aware metadata for the History API router.
// index.html holds crawler-friendly defaults; this keeps tags accurate on SPA navigation.

const SITE_URL = 'https://only4.fun';

const pages = {
  '/': {
    title: 'Only4fun — Free Static Site Hosting in Seconds',
    description:
      'Deploy static sites from a GitHub repo or ZIP with no signup. Get a free yoursite.only4.fun address, instant SSL, and one-click launch.',
  },
  '/instant': {
    title: 'Instant Launch — Deploy a Static Site Without Signing Up | Only4fun',
    description:
      'Paste a public GitHub link or drop a ZIP and deploy static HTML, CSS and JavaScript instantly. No account, no config, no credit card.',
  },
  '/launchpad': {
    title: 'Your Projects — Only4fun Launchpad',
    description:
      'Manage your deployed sites, environment variables, custom domains, traffic stats and deployment history in the Only4fun launchpad.',
    private: true,
  },
  '/deploy': {
    title: 'Deploy New — GitHub or ZIP | Only4fun',
    description:
      'Deploy a new project from a GitHub repository or a ZIP file, choose a subdomain, and build frameworks with the full toolchain.',
    private: true,
  },
  '/docs': {
    title: 'Docs — Only4fun',
    description:
      'Guides and answers for deploying static sites and framework projects to Only4fun, from first upload to custom domains.',
  },
  '/status': {
    title: 'Status — Only4fun',
    description:
      'Check the current health of Only4fun deployments, DNS, and the build pipeline.',
  },
  '/support': {
    title: 'Support — Only4fun',
    description:
      'Need a hand? Find help for deploys, subdomains, custom domains and build errors on Only4fun.',
  },
  '/terms': {
    title: 'Terms And Conditions — Only4fun',
    description:
      'Read the terms and conditions for using the Only4fun static site hosting service.',
  },
  '/privacy': {
    title: 'Privacy Policy — Only4fun',
    description:
      'Learn what data Only4fun collects, how session and deployment data is handled, and your choices.',
  },
  '/project/': {
    title: 'Project Details — Only4fun',
    description:
      'Inspect deployment history, traffic, environment variables and settings for an Only4fun project.',
    private: true,
  },
};

function setContent(id, value) {
  const el = document.getElementById(id);
  if (el) el.setAttribute('content', value);
}

export function updateSeo(route) {
  if (typeof document === 'undefined') return;

  const path = route === '/' ? '/' : route.replace(/\/+$/, '') || '/';
  const page = pages[path] ?? (path.startsWith('/project/') ? pages['/project/'] : pages['/']);
  const url = path === '/' ? `${SITE_URL}/` : `${SITE_URL}${path}`;

  document.title = page.title;
  setContent('meta-description', page.description);
  setContent('og-title', page.title);
  setContent('og-description', page.description);
  setContent('twitter-title', page.title);
  setContent('twitter-description', page.description);
  setContent('og-url', url);
  setContent(
    'meta-robots',
    page.private ? 'noindex, nofollow' : 'index, follow, max-snippet:-1, max-image-preview:large, max-video-preview:-1',
  );

  const canonical = document.getElementById('canonical');
  if (canonical) canonical.setAttribute('href', url);

  document.querySelectorAll('link[rel="alternate"][hreflang]').forEach((link) => {
    link.setAttribute('href', url);
  });
}
