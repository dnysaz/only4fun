<script>
  import Dots from './Dots.svelte';
  import { api } from './api.js';
  import { session } from './session.svelte.js';
  import { getRoute, navigate } from './router.svelte.js';

  let { onSignOut } = $props();

  let initials = $derived((session.user?.username ?? '?').slice(0, 2).toUpperCase());
  let route = $derived(getRoute());
  let hash = $state(window.location.hash.slice(1));
  let scrolled = $state(false);
  let isMobile = $state(window.matchMedia('(max-width: 700px)').matches);
  let drawerOpen = $state(false);

  $effect(() => {
    const query = window.matchMedia('(max-width: 700px)');
    function syncViewport() {
      isMobile = query.matches;
      if (!query.matches) drawerOpen = false;
    }
    function syncHash() {
      hash = window.location.hash.slice(1);
    }
    function syncScroll() {
      scrolled = window.scrollY > 8;
    }
    syncViewport();
    syncScroll();
    query.addEventListener('change', syncViewport);
    window.addEventListener('hashchange', syncHash);
    window.addEventListener('scroll', syncScroll, { passive: true });
    return () => {
      query.removeEventListener('change', syncViewport);
      window.removeEventListener('hashchange', syncHash);
      window.removeEventListener('scroll', syncScroll);
    };
  });

  $effect(() => {
    if (!drawerOpen) return;
    const previousOverflow = document.body.style.overflow;
    function onKeydown(event) {
      if (event.key === 'Escape') drawerOpen = false;
    }
    document.body.style.overflow = 'hidden';
    window.addEventListener('keydown', onKeydown);
    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener('keydown', onKeydown);
    };
  });

  function isActive(href) {
    if (href === '/instant') return route === '/instant';
    if (href === '/launchpad') {
      return route === '/launchpad' || route === '/deploy' || route.startsWith('/project/');
    }
    return false;
  }

  function isSectionActive(id) {
    return route === '/' && hash === id;
  }

  function goDashboard(event) {
    event.preventDefault();
    navigate('/launchpad');
  }

  function handleAvatar(event) {
    if (!isMobile) {
      goDashboard(event);
      return;
    }
    event.preventDefault();
    drawerOpen = !drawerOpen;
  }

  function closeDrawer() {
    drawerOpen = false;
  }

  function signOutFromDrawer() {
    closeDrawer();
    onSignOut?.();
  }

  function goInstant(event) {
    event.preventDefault();
    navigate('/instant');
  }

  // Section link tetap berfungsi saat navbar dipakai di halaman /instant.
  function goSection(event, id) {
    if (getRoute() === '/') return;
    event.preventDefault();
    navigate('/');
    requestAnimationFrame(() =>
      requestAnimationFrame(() => {
        document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' });
        hash = id;
        history.replaceState({}, '', `#${id}`);
      }),
    );
  }
</script>

<nav class:scrolled>
  <a class="logo" href="/" onclick={(e) => { e.preventDefault(); navigate('/'); }} aria-label="Only4fun home"><Dots /><span class="word">Only4<span class="fun">fun</span></span></a>
  <ul>
    <li><a href="#how" class:active={isSectionActive('how')} onclick={(event) => goSection(event, 'how')}>How it works</a></li>
    <li><a href="#features" class:active={isSectionActive('features')} onclick={(event) => goSection(event, 'features')}>Features</a></li>
    <li><a href="/instant" class:active={isActive('/instant')} onclick={goInstant}>Instant Launch <span class="nav-arrow" aria-hidden="true"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M7 17 17 7" /><path d="M8 7h9v9" /></svg></span></a></li>
    <li class="actions">
      {#if session.user}
        <a
          class="avatar"
          href="/launchpad"
          onclick={handleAvatar}
          aria-label={isMobile ? 'Open account menu' : `Open launchpad as @${session.user.username}`}
          title={`@${session.user.username}`}
        >
          {#if session.user.avatar_url}
            <img src={session.user.avatar_url} alt="" width="34" height="34" />
          {:else}
            <span>{initials}</span>
          {/if}
        </a>
        <a class="btn launchpad-btn" href="/launchpad" onclick={goDashboard}>Launchpad</a>
        {#if onSignOut}
          <button class="signout" type="button" onclick={onSignOut} aria-label="Sign out" title="Sign out">
            <svg class="signout-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
              <path d="m16 17 5-5-5-5" />
              <path d="M21 12H9" />
            </svg>
            <span class="signout-label">Sign out</span>
          </button>
        {/if}
      {:else}
        <a class="btn" href={api.loginUrl('/launchpad')}>Get started</a>
      {/if}
    </li>
  </ul>
</nav>

<button
  class="drawer-backdrop"
  class:open={drawerOpen}
  type="button"
  aria-label="Close account menu"
  tabindex={drawerOpen ? 0 : -1}
  onclick={closeDrawer}
></button>

<div
  class="account-drawer"
  class:open={drawerOpen}
  role="dialog"
  aria-modal="true"
  aria-label="Account menu"
  aria-hidden={!drawerOpen}
>
  <div class="drawer-handle" aria-hidden="true"></div>
  <div class="drawer-user">
    <span class="avatar">
      {#if session.user?.avatar_url}
        <img src={session.user.avatar_url} alt="" width="42" height="42" />
      {:else}
        <span>{initials}</span>
      {/if}
    </span>
    <div>
      <b>@{session.user?.username ?? 'you'}</b>
      <span>Only4fun account</span>
    </div>
  </div>

  {#if onSignOut}
    <button class="drawer-item" type="button" onclick={signOutFromDrawer}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
        <path d="m16 17 5-5-5-5" />
        <path d="M21 12H9" />
      </svg>
      <span>Sign out</span>
    </button>
  {/if}
</div>

<style>
  nav {
    position: fixed;
    /* Nav mengambang dengan ujung kiri-kanan melengkung (pill). */
    top: calc(env(safe-area-inset-top, 0px) + 12px);
    left: var(--page-x);
    right: var(--page-x);
    z-index: 50;
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: calc(var(--nav-h) - 24px);
    padding: 0 28px;
    background-color: var(--bg);
    border: 1px solid transparent;
    border-radius: 999px;
    transition: border-color 160ms ease;
  }

  nav.scrolled {
    border-color: var(--line);
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 10px;
    font-weight: 800;
    font-size: 28px;
    letter-spacing: -0.045em;
    white-space: nowrap;
    text-decoration: none;
  }

  .logo .word {
    display: inline-block;
  }

  .logo .fun {
    color: var(--blue);
  }

  nav ul {
    display: flex;
    gap: 32px;
    list-style: none;
    margin: 0;
    padding: 0;
    align-items: center;
  }

  nav a {
    text-decoration: none;
    font-weight: 500;
    color: var(--muted);
  }

  nav a:hover {
    color: var(--ink);
  }

  nav a.active {
    color: var(--blue);
    text-decoration: underline;
    text-decoration-thickness: 2px;
    text-underline-offset: 6px;
  }

  nav a.active:hover {
    color: var(--blue);
  }

  .nav-arrow {
    display: inline-flex;
    align-items: center;
    margin-left: 5px;
    transition: transform 160ms ease;
  }

  .nav-arrow svg {
    width: 15px;
    height: 15px;
  }

  nav a:hover .nav-arrow {
    transform: translate(2px, -2px);
  }

  .actions {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .avatar {
    display: grid;
    place-items: center;
    width: 34px;
    height: 34px;
    border-radius: 50%;
    overflow: hidden;
    background: var(--soft);
    border: 1px solid var(--line);
    font: 600 13px/1 'Gabarito', sans-serif;
    color: var(--ink);
  }

  .avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  nav a.btn {
    color: #fff;
    padding: 12px 24px;
    font-size: 17px;
    white-space: nowrap;
  }

  nav button.signout {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    background: transparent;
    border: 1.5px solid var(--muted);
    border-radius: 999px;
    padding: 12px 24px;
    font: 600 17px/1 'Gabarito', sans-serif;
    color: var(--ink);
    white-space: nowrap;
    cursor: pointer;
  }

  .signout-icon {
    display: none;
    width: 18px;
    height: 18px;
  }

  nav button.signout:hover {
    background: transparent !important;
    border-color: var(--ink) !important;
    color: var(--ink) !important;
  }

  .drawer-backdrop {
    position: fixed;
    inset: 0;
    z-index: 59;
    padding: 0;
    border: 0;
    background: rgba(32, 33, 36, 0.45);
    opacity: 0;
    visibility: hidden;
    transition: opacity 220ms ease, visibility 220ms ease;
  }

  .drawer-backdrop.open {
    opacity: 1;
    visibility: visible;
  }

  .account-drawer {
    position: fixed;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 60;
    height: 50vh;
    max-height: 420px;
    padding: 10px var(--page-x) calc(22px + env(safe-area-inset-bottom, 0px));
    background: var(--bg);
    border-top: 1px solid var(--line);
    border-radius: 24px 24px 0 0;
    transform: translateY(100%);
    visibility: hidden;
    transition: transform 260ms cubic-bezier(0.22, 0.8, 0.24, 1), visibility 260ms ease;
  }

  .account-drawer.open {
    transform: translateY(0);
    visibility: visible;
  }

  .drawer-handle {
    width: 42px;
    height: 4px;
    margin: 0 auto 18px;
    border-radius: 999px;
    background: var(--line);
  }

  .drawer-user {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-bottom: 22px;
  }

  .drawer-user .avatar {
    width: 42px;
    height: 42px;
    font-size: 15px;
  }

  .drawer-user b {
    display: block;
    color: var(--ink);
    font-size: 17px;
  }

  .drawer-user div span {
    display: block;
    margin-top: 3px;
    color: var(--muted);
    font-size: 14px;
  }

  button.drawer-item {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;
    padding: 16px 0;
    border: 0;
    border-top: 1px solid var(--line);
    border-radius: 0;
    background: transparent;
    color: var(--ink);
    font: 600 17px/1 'Gabarito', sans-serif;
    text-align: left;
  }

  button.drawer-item svg {
    width: 20px;
    height: 20px;
  }

  @media (min-width: 701px) {
    .drawer-backdrop,
    .account-drawer {
      display: none;
    }
  }

  @media (max-width: 900px) {
    nav ul li:not(.actions) {
      display: none;
    }
  }

  @media (max-width: 700px) {
    nav {
      padding: 0 18px;
    }

    .logo {
      font-size: 24px;
      gap: 8px;
    }

    .actions {
      gap: 8px;
    }

    nav a.btn {
      padding: 10px 18px;
      font-size: 15px;
    }

    nav a.launchpad-btn {
      display: none;
    }

    nav button.signout {
      display: none;
    }
  }
</style>
