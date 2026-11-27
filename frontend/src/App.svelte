<script>
  import { onMount } from 'svelte';
  import Nav from './lib/Nav.svelte';
  import Hero from './lib/Hero.svelte';
  import DeployBand from './lib/DeployBand.svelte';
  import Statement from './lib/Statement.svelte';
  import Features from './lib/Features.svelte';
  import Cta from './lib/Cta.svelte';
  import Footer from './lib/Footer.svelte';
  import Dashboard from './lib/Dashboard.svelte';
  import Deploy from './lib/Deploy.svelte';
  import ProjectDetail from './lib/ProjectDetail.svelte';
  import InstantDeploy from './lib/InstantDeploy.svelte';
  import InfoPage from './lib/InfoPage.svelte';
  import { getRoute, navigation } from './lib/router.svelte.js';
  import { Loading } from './lib/loading.ts';
  import { getInfoPage } from './lib/infoPages.js';
  import { loadSession, revalidateSession } from './lib/session.svelte.js';
  import { logout } from './lib/authActions.js';
  import { updateSeo } from './lib/seo.js';

  onMount(() => {
    loadSession();

    // Tab yang dibiarkan semalaman atau perpindah halaman harus memastikan
    // sesinya masih valid sebelum launchpad/deploy dipakai.
    function onVisible() {
      if (document.visibilityState === 'visible') revalidateSession();
    }
    const idleTimer = window.setInterval(() => {
      if (!document.hidden) revalidateSession();
    }, 5 * 60 * 1000);

    document.addEventListener('visibilitychange', onVisible);
    return () => {
      window.clearInterval(idleTimer);
      document.removeEventListener('visibilitychange', onVisible);
    };
  });

  let lastRoute = '';
  $effect(() => {
    const route = getRoute();
    updateSeo(route);
    if (lastRoute && route !== lastRoute) revalidateSession();
    lastRoute = route;
  });
</script>

{#if getRoute() === '/launchpad'}
  <Dashboard />
{:else if getRoute() === '/deploy'}
  <Deploy />
{:else if getRoute() === '/instant'}
  <InstantDeploy />
{:else if getRoute().startsWith('/project/')}
  <ProjectDetail />
{:else if getInfoPage(getRoute())}
  <div class="wrap">
    <Nav onSignOut={logout} />
  </div>
  <InfoPage page={getRoute()} />
{:else}
  <div class="wrap">
    <Nav onSignOut={logout} />
  </div>

  <Hero />

  <DeployBand />
  <Statement />
  <Features />
  <Cta />
  <Footer />
{/if}

{#if navigation.loading}
  <div class="route-loading">
    <Loading label="loading page…" size={16} inline />
  </div>
{/if}

<style>
  .route-loading {
    position: fixed;
    left: 50%;
    bottom: calc(22px + env(safe-area-inset-bottom, 0px));
    z-index: 80;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    padding: 11px 18px;
    border: 1px solid var(--line);
    border-radius: 999px;
    background: var(--bg);
    font-size: 13px;
    color: var(--muted);
  }
</style>
