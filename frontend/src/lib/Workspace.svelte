<script>
  import { onMount } from 'svelte';
  import Dots from './Dots.svelte';
  import Nav from './Nav.svelte';
  import Footer from './Footer.svelte';
  import { api } from './api.js';
  import { session, signOut } from './session.svelte.js';
  import { navigate } from './router.svelte.js';
  import { workspace, ensureWorkspace, reload, resetWorkspace } from './workspace.svelte.js';

  let { view = 'projects', compact = false, crumb = '', children } = $props();

  let total = $derived(workspace.projects.length);
  let live = $derived(workspace.projects.filter((p) => p.status === 'deployed').length);
  let views = $derived(workspace.projects.reduce((sum, p) => sum + (p.views ?? 0), 0));
  let viewLabel = $derived(new Intl.NumberFormat().format(views));
  let building = $derived(
    workspace.projects.some((p) => p.status === 'pending' || p.status === 'building'),
  );
  let loginNext = $derived(
    `${view === 'deploy' ? '/deploy' : '/launchpad'}${typeof window === 'undefined' ? '' : window.location.search}`,
  );
  onMount(() => {
    ensureWorkspace();
  });

  // Selama ada build berjalan, cek status tiap 5 detik.
  $effect(() => {
    if (!building || !session.user) return;
    const timer = setInterval(() => reload({ quiet: true }), 5000);
    return () => clearInterval(timer);
  });

  function go(event, path) {
    event.preventDefault();
    navigate(path);
  }

  async function logout() {
    await signOut();
    resetWorkspace();
    navigate('/');
  }
</script>

<div class="page wrap workspace">
  <Nav onSignOut={logout} />

  {#if session.loading || (workspace.loading && !workspace.projects.length)}
    <section class="splash">
      <span class="loading"><Dots size={30} /></span>
      <p>waking up your launchpad…</p>
    </section>
  {:else if !session.user}
    <section class="gate">
      <div class="copy">
        <p class="eyebrow"><span class="pulse" aria-hidden="true"></span>Free forever · no credit card</p>
        <h1>Your projects,<br /><span class="hl">all in one place.</span></h1>
        <p class="lead">
          Connect GitHub once. Every project you ship shows up here — live status, views and one-click
          redeploys.
        </p>
        <a class="btn" href={api.loginUrl(loginNext)}>
          Continue with GitHub
          <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M7 17L17 7" /><path d="M8 7h9v9" />
          </svg>
        </a>
        {#if workspace.error}
          <p class="status bad">{workspace.error}</p>
        {/if}
      </div>
    </section>
  {:else if compact}
    <nav class="backnav" aria-label="Breadcrumb">
      <a href="/launchpad" onclick={(e) => go(e, '/launchpad')}>Your projects</a>
      <span>/ {crumb}</span>
    </nav>

    {@render children?.()}
  {:else}
    <section class="intro">
      <div class="copy">
        <p class="eyebrow"><span class="pulse" aria-hidden="true"></span>Launchpad</p>
        <h1>Hey <span class="hl">@{session.user.username}</span>.<br />Ship something fun.</h1>
        <p class="lead">
          Paste a repo or drop a ZIP — your site goes live at
          <span class="domain">yoursite.only4.fun</span>.
        </p>
      </div>
      <ul class="stats">
        <li class="blue"><b>{total}</b> projects</li>
        <li class="green"><b>{live}</b> live</li>
        <li class="yellow"><b>{viewLabel}</b> views</li>
      </ul>
    </section>

    <nav class="subnav" aria-label="Launchpad pages">
      <a
        href="/launchpad"
        class:active={view === 'projects'}
        aria-current={view === 'projects' ? 'page' : undefined}
        onclick={(e) => go(e, '/launchpad')}
      >Your projects</a>
      <a
        href="/deploy"
        class:active={view === 'deploy'}
        aria-current={view === 'deploy' ? 'page' : undefined}
        onclick={(e) => go(e, '/deploy')}
      >Deploy new</a>
      {#if building}
        <span class="building"><i></i>building…</span>
      {/if}
    </nav>

    {@render children?.()}
  {/if}
</div>

<Footer />

<style>
  .page {
    min-height: 70vh;
    padding-bottom: clamp(48px, 7vw, 96px);
  }

  .splash {
    display: grid;
    place-items: center;
    gap: 20px;
    padding: 130px 0;
  }

  .loading {
    animation: wobble 1.6s ease-in-out infinite;
  }

  @keyframes wobble {
    0%, 100% { transform: rotate(-8deg) scale(1); }
    50% { transform: rotate(8deg) scale(1.08); }
  }

  .splash p {
    color: var(--muted);
    font: 500 16px/1 'DM Mono', monospace;
  }

  /* Sama seperti homepage: pill border tipis + dot, highlight kuning, tombol biru. */
  .eyebrow {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    margin: 0 0 26px;
    padding: 8px 16px 8px 14px;
    border: 1px solid var(--line);
    border-radius: 999px;
    font: 500 14px/1 'DM Mono', monospace;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--muted);
  }

  .pulse {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--green);
  }

  h1 {
    margin: 0;
    font-size: clamp(48px, 7.4vw, 96px);
    line-height: 0.95;
    letter-spacing: -0.045em;
    font-weight: 800;
    text-wrap: balance;
  }

  .hl {
    background: linear-gradient(transparent 62%, rgba(251, 188, 4, 0.5) 62%);
    padding: 0 0.04em;
  }

  .lead {
    max-width: 34em;
    margin: 22px 0 0;
    color: var(--muted);
    font-size: clamp(19px, 1.7vw, 24px);
    line-height: 1.45;
  }

  .domain {
    color: var(--ink);
    text-decoration: underline;
    text-decoration-thickness: 2px;
    text-underline-offset: 0.18em;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    gap: 9px;
    margin-top: 30px;
    white-space: nowrap;
    padding: 17px 26px;
    font-size: 18px;
  }

  .status {
    margin: 14px 0 0;
    font-size: 17px;
    color: var(--muted);
  }

  .status.bad { color: var(--red); }

  .gate {
    display: grid;
    grid-template-columns: minmax(0, 760px);
    align-items: start;
    /* full bleed supaya area ini sampai ke tepi viewport, bukan hanya
       selebar area .wrap */
    width: 100vw;
    margin-left: calc(50% - 50vw);
    padding: 64px var(--page-x) 72px;
    /* offset padding bawah .page supaya jarak ke footer sama seperti homepage */
    margin-bottom: clamp(-96px, -7vw, -48px);
    overflow-x: clip;
  }

  .intro {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 32px;
    flex-wrap: wrap;
    padding: 64px 0 40px;
  }

  .stats {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    list-style: none;
    margin: 0 0 6px;
    padding: 0;
  }

  .stats li {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    border: 1px solid var(--line);
    border-radius: 999px;
    padding: 10px 20px 10px 14px;
    font-weight: 600;
    font-size: 19px;
  }

  .stats li::before {
    content: '';
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: var(--blue);
  }

  .stats li.green::before { background: var(--green); }
  .stats li.yellow::before { background: var(--yellow); }

  .stats b {
    font-weight: 800;
  }

  .backnav {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 46px 0 26px;
    color: var(--muted);
    font: 500 18px/1 'Gabarito', sans-serif;
  }

  .backnav a {
    color: var(--ink);
    text-decoration: none;
    border-bottom: 2px solid var(--yellow);
    padding-bottom: 3px;
    font-weight: 600;
  }

  .backnav a:hover {
    color: var(--ink);
  }

  .subnav {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
    padding-bottom: 36px;
  }

  .subnav a {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    padding: 11px 20px 11px 16px;
    border: 1px solid var(--line);
    border-radius: 999px;
    color: var(--muted);
    font: 600 16px/1 'Gabarito', sans-serif;
    text-decoration: none;
  }

  .subnav a::before {
    content: '';
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: var(--line);
  }

  .subnav a:hover {
    color: var(--ink);
    background: var(--soft);
  }

  .subnav a.active {
    color: var(--blue);
    border-color: var(--blue);
    background: color-mix(in srgb, var(--blue) 7%, var(--bg));
  }

  .subnav a.active::before {
    background: var(--blue);
  }

  .building {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    margin-left: auto;
    color: var(--muted);
    font: 500 14px/1 'DM Mono', monospace;
  }

  .building i {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--yellow);
  }

  @media (max-width: 900px) {
    .gate,
    .intro {
      padding-top: 32px;
    }

    .building { margin-left: 0; }
  }

  @media (prefers-reduced-motion: reduce) {
    .loading {
      animation: none;
    }
  }
</style>
