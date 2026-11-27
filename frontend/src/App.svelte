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
  import { getRoute } from './lib/router.svelte.js';
  import { getInfoPage } from './lib/infoPages.js';
  import { loadSession, revalidateSession } from './lib/session.svelte.js';
  import { updateSeo } from './lib/seo.js';

  onMount(() => {
    loadSession();

    // Tab yang dibiarkan semalaman harus memastikan sesinya masih valid
    // sebelum halaman launchpad/deploy dipakai.
    function onVisible() {
      if (document.visibilityState === 'visible') revalidateSession();
    }
    document.addEventListener('visibilitychange', onVisible);
    return () => document.removeEventListener('visibilitychange', onVisible);
  });

  $effect(() => {
    updateSeo(getRoute());
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
    <Nav />
  </div>
  <InfoPage page={getRoute()} />
{:else}
  <div class="wrap">
    <Nav />
  </div>

  <Hero />

  <DeployBand />
  <Statement />
  <Features />
  <Cta />
  <Footer />
{/if}
