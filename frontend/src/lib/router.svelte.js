// Router minimal berbasis History API (tanpa dependency).

function currentPath() {
  let path = window.location.pathname.replace(/\/+$/, '');
  if (path === '') path = '/';
  // /dashboard adalah nama lama; sekarang halaman ini bernama /launchpad.
  if (path === '/dashboard') {
    path = '/launchpad';
    history.replaceState({}, '', `/launchpad${window.location.search}`);
  }
  return path;
}

let path = $state(currentPath());
let query = $state(new URLSearchParams(window.location.search));

function sync() {
  path = currentPath();
  query = new URLSearchParams(window.location.search);
}

if (typeof window !== 'undefined') {
  window.addEventListener('popstate', sync);
}

export function getRoute() {
  return path;
}

export function getQuery() {
  return query;
}

export function navigate(to, { replace = false } = {}) {
  const target = to.startsWith('/') ? to : `/${to}`;
  // Sertakan hash agar navigate('/') benar-benar menghapus #features/#how.
  const current = window.location.pathname + window.location.search + window.location.hash;
  if (target === current) return;
  if (replace) history.replaceState({}, '', target);
  else history.pushState({}, '', target);
  sync();
  window.scrollTo({ top: 0, behavior: 'instant' });
}
