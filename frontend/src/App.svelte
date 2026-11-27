<script>
  import { onMount } from 'svelte';
  import Nav from './lib/Nav.svelte';
  import Hero from './lib/Hero.svelte';
  import DeployBand from './lib/DeployBand.svelte';
  import Statement from './lib/Statement.svelte';
  import Features from './lib/Features.svelte';
  import Cta from './lib/Cta.svelte';
  import Footer from './lib/Footer.svelte';
  import InfoPage from './lib/InfoPage.svelte';
  import { getRoute, navigation } from './lib/router.svelte.js';
  import { Loading } from './lib/loading.ts';
  import { getInfoPage } from './lib/infoPages.js';
  import { loadSession, revalidateSession } from './lib/session.svelte.js';
  import { logout } from './lib/authActions.js';
  import { updateSeo } from './lib/seo.js';
  import { watchPageVisibility } from './lib/visibility.svelte.js';

  // Halaman workspace/component berat dimuat sesuai route. Homepage/info
  // tetap statis supaya first paint cepat.
  const routeLoaders = {
    '/launchpad': () => import('./lib/Dashboard.svelte'),
    '/deploy': () => import('./lib/Deploy.svelte'),
    '/instant': () => import('./lib/InstantDeploy.svelte'),
  };
  let ActiveRoute = $state(null);
  let loadToken = 0;

  let route = $derived(getRoute());
  let workspaceRoute = $derived(
    route === '/launchpad' ||
      route === '/deploy' ||
      route === '/instant' ||
      route.startsWith('/project/'),
  );
  let infoPage = $derived(getInfoPage(route));

  $effect(() => {
    const current = route;
    const loader =
      routeLoaders[current] ??
      (current.startsWith('/project/') ? () => import('./lib/ProjectDetail.svelte') : null);
    const token = ++loadToken;
    let cancelled = false;

    if (!loader) {
      ActiveRoute = null;
      return;
    }

    ActiveRoute = null;
    loader()
      .then((module) => {
        if (cancelled || token !== loadToken) return;
        ActiveRoute = module.default;
      })
      .catch((error) => {
        if (cancelled || token !== loadToken) return;
        console.error('Failed to load route component', error);
        ActiveRoute = null;
      });

    return () => {
      cancelled = true;
    };
  });

  onMount(() => {
    const stopVisibility = watchPageVisibility();
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
      stopVisibility();
      window.clearInterval(idleTimer);
      document.removeEventListener('visibilitychange', onVisible);
    };
  });

  let lastRoute = '';
  $effect(() => {
    updateSeo(route);
    if (lastRoute && route !== lastRoute) revalidateSession();
    lastRoute = route;
  });
</script>

{#if workspaceRoute}
  {#if ActiveRoute}
    <ActiveRoute />
  {:else}
    <div class="route-splash"><Loading label="loading page…" size={30} /></div>
  {/if}
{:else if infoPage}
  <div class="wrap">
    <Nav onSignOut={logout} />
  </div>
  <InfoPage page={route} />
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
  .route-splash {
    display: grid;
    place-items: center;
    min-height: 70vh;
  }

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
