// Helper untuk menampilkan file sebagai hierarchy dan mendeteksi project
// framework. Server tetap menjadi validator utama saat deploy.

const MAX_TREE_FILES = 400;

const FRAMEWORK_NAMES = new Set([
  'package.json',
  'package-lock.json',
  'yarn.lock',
  'pnpm-lock.yaml',
  'bun.lock',
  'bun.lockb',
  'deno.json',
  'tsconfig.json',
  'next.config.js',
  'next.config.mjs',
  'next.config.ts',
  'vite.config.js',
  'vite.config.mjs',
  'vite.config.ts',
  'nuxt.config.js',
  'nuxt.config.mjs',
  'nuxt.config.ts',
  'astro.config.js',
  'astro.config.mjs',
  'astro.config.ts',
  'svelte.config.js',
  'angular.json',
  'gatsby-config.js',
  'gatsby-config.ts',
  'webpack.config.js',
  'webpack.config.mjs',
  'rollup.config.js',
  'server.js',
  'server.mjs',
  'server.cjs',
  '.next',
  '.nuxt',
  '.svelte-kit',
  '.astro',
  'node_modules',
]);

const FRAMEWORK_EXTS = new Set(['.vue', '.svelte', '.jsx', '.tsx', '.ts']);

// Output build framework sering tidak punya package.json atau sumber .jsx,
// jadi perlu dikenali dari struktur hasil build-nya.
const FRAMEWORK_BUILD_PATHS = [
  /(^|\/)_next(\/|$)/i,
  /(^|\/)__next(\/|$)/i,
  /(^|\/)_astro(\/|$)/i,
  /(^|\/)\.output(\/|$)/i,
  /(^|\/)asset-manifest\.json$/i,
  /(^|\/)static\/js\/[^/]+\.chunk\.js$/i,
  /(^|\/)assets\/[^/]+[-.][a-f0-9]{6,}\.(?:js|mjs)$/i,
];

// Sama seperti stripSingleRoot di backend: bila semua file berada dalam satu
// folder dan folder itu berisi index.html, folder tersebut jadi root tampilan.
export function normalizeTreePaths(paths) {
  const cleaned = (Array.isArray(paths) ? paths : [])
    .map((raw) =>
      String(raw)
        .replace(/\\/g, '/')
        .replace(/^\.\//, '')
        .trim(),
    )
    .filter((path) => path && !path.startsWith('/'))
    .filter((path) => !path.split('/').some((part) => !part || part === '.' || part === '..'));

  if (cleaned.includes('index.html')) return cleaned;

  const roots = new Set(cleaned.map((path) => path.split('/')[0]));
  if (roots.size === 1) {
    const [root] = roots;
    if (cleaned.includes(`${root}/index.html`)) {
      return cleaned
        .filter((path) => path !== root)
        .map((path) => path.slice(root.length + 1))
        .filter(Boolean);
    }
  }
  return cleaned;
}

export function buildTreeRows(paths) {
  const normalized = normalizeTreePaths(paths);
  const extra = Math.max(0, normalized.length - MAX_TREE_FILES);
  const children = new Map();

  const bucket = (parent) => {
    if (!children.has(parent)) children.set(parent, new Map());
    return children.get(parent);
  };
  const add = (parent, node) => {
    const map = bucket(parent);
    const key = `${node.kind}:${node.name.toLowerCase()}`;
    if (!map.has(key)) map.set(key, node);
  };

  for (const path of normalized.slice(0, MAX_TREE_FILES)) {
    const parts = path.split('/');
    let parent = '';
    for (let i = 0; i < parts.length - 1; i++) {
      const name = parts[i];
      const full = parent ? `${parent}/${name}` : name;
      add(parent, { name, path: full, kind: 'folder' });
      parent = full;
    }
    add(parent, { name: parts[parts.length - 1], path, kind: 'file' });
  }

  const rows = [];
  const walk = (parent, depth) => {
    const items = [...(children.get(parent)?.values() ?? [])].sort((a, b) => {
      if (a.kind !== b.kind) return a.kind === 'folder' ? -1 : 1;
      return a.name.localeCompare(b.name, undefined, { sensitivity: 'base' });
    });
    for (const item of items) {
      rows.push({ ...item, depth });
      if (item.kind === 'folder') walk(item.path, depth + 1);
    }
  };
  walk('', 0);

  if (extra > 0) {
    rows.push({ name: `+${extra} more files`, path: '__more__', kind: 'more', depth: 0 });
  }
  return rows;
}

export function detectFramework(paths, signals = []) {
  if (Array.isArray(signals) && signals.length > 0) return true;
  return normalizeTreePaths(paths).some((path) => {
    const parts = path.split('/');
    if (parts.some((part) => FRAMEWORK_NAMES.has(part.toLowerCase()))) return true;
    if (FRAMEWORK_BUILD_PATHS.some((pattern) => pattern.test(path))) return true;
    const dot = path.lastIndexOf('.');
    return dot >= 0 && FRAMEWORK_EXTS.has(path.slice(dot).toLowerCase());
  });
}
