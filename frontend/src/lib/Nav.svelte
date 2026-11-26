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

  $effect(() => {
    function syncHash() {
      hash = window.location.hash.slice(1);
    }
    function syncScroll() {
      scrolled = window.scrollY > 8;
    }
    syncScroll();
    window.addEventListener('hashchange', syncHash);
    window.addEventListener('scroll', syncScroll, { passive: true });
    return () => {
      window.removeEventListener('hashchange', syncHash);
      window.removeEventListener('scroll', syncScroll);
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
    <li><a href="/instant" class:active={isActive('/instant')} onclick={goInstant}>Instant Launch <span class="nav-arrow" aria-hidden="true">↗</span></a></li>
    <li class="actions">
      {#if session.user}
        <a class="avatar" href="/launchpad" onclick={goDashboard} title={`@${session.user.username}`}>
          {#if session.user.avatar_url}
            <img src={session.user.avatar_url} alt="" width="34" height="34" />
          {:else}
            <span>{initials}</span>
          {/if}
        </a>
        <a class="btn" href="/launchpad" onclick={goDashboard}>Launchpad</a>
        {#if onSignOut}
          <button class="signout" type="button" onclick={onSignOut}>Sign out</button>
        {/if}
      {:else}
        <a class="btn" href={api.loginUrl('/launchpad')}>Get started</a>
      {/if}
    </li>
  </ul>
</nav>

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
    display: inline-block;
    margin-left: 5px;
    font-size: 0.95em;
    line-height: 1;
    transition: transform 160ms ease;
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
  }

  nav button.signout {
    background: transparent;
    border: 1px solid var(--line);
    border-radius: 999px;
    padding: 11px 20px;
    font: 600 15px/1 'Gabarito', sans-serif;
    color: var(--ink);
    cursor: pointer;
  }

  nav button.signout:hover {
    border-color: var(--ink);
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

    nav a.btn {
      padding: 10px 18px;
      font-size: 15px;
    }

    nav button.signout {
      padding: 9px 14px;
      font-size: 14px;
    }
  }
</style>
