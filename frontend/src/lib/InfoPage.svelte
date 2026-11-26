<script>
  import Footer from './Footer.svelte';
  import { getInfoPage } from './infoPages.js';
  import { navigate } from './router.svelte.js';

  let { page } = $props();
  let data = $derived(getInfoPage(page));

  function go(event, href) {
    event.preventDefault();
    navigate(href);
  }
</script>

{#if data}
  <section class="info wrap">
    <header class="masthead">
      <p class="eyebrow">{data.eyebrow}</p>
      <h1>{data.title} <span class="hl">{data.accent}</span></h1>
      <p class="lead">{data.lead}</p>
      <div class="meta">
        <span>{data.updated}</span>
        {#if data.badge}
          <span class="badge"><i aria-hidden="true"></i>{data.badge}</span>
        {/if}
      </div>
    </header>

    <div class="body-grid">
      <aside class="index" aria-label="On this page">
        <p class="index-title">On this page</p>
        <ol>
          {#each data.sections as section (section.id)}
            <li><a href={`#${section.id}`}>{section.heading}</a></li>
          {/each}
        </ol>
      </aside>

      <div class="sections">
        {#each data.sections as section (section.id)}
          <article id={section.id}>
            <h2>{section.heading}</h2>
            <p>{section.body}</p>

            {#if data.status && section.id === 'now'}
              <div class="status-list">
                {#each data.status as item (item.name)}
                  <div class="status-row">
                    <span><i aria-hidden="true"></i>{item.name}</span>
                    <span class="state">{item.state}</span>
                  </div>
                {/each}
              </div>
            {/if}

            {#if section.items}
              <ul>
                {#each section.items as item}
                  <li>{item}</li>
                {/each}
              </ul>
            {/if}
          </article>
        {/each}
      </div>
    </div>

    <div class="cta">
      <div>
        <h2>{data.cta.title}</h2>
        <p>{data.cta.body}</p>
      </div>
      <a class="btn" href={data.cta.href} onclick={(event) => go(event, data.cta.href)}>{data.cta.label}</a>
    </div>
  </section>

  <Footer />
{/if}

<style>
  .info {
    /* Jangan timpa padding horizontal dari .wrap. */
    padding-block: 64px clamp(72px, 8vw, 104px);
  }

  .masthead {
    max-width: none;
  }

  .eyebrow {
    display: inline-flex;
    align-items: center;
    margin: 0 0 26px;
    padding: 8px 16px;
    border: 1px solid var(--line);
    border-radius: 999px;
    font: 500 14px/1 'DM Mono', monospace;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--muted);
  }

  h1 {
    font-size: clamp(52px, 7.8vw, 104px);
    line-height: 0.95;
    letter-spacing: -0.045em;
    font-weight: 800;
    margin: 0;
    white-space: nowrap;
  }

  .hl {
    background: linear-gradient(transparent 62%, rgba(251, 188, 4, 0.55) 62%);
    padding: 0 0.04em;
  }

  .lead {
    margin: 32px 0 0;
    max-width: 46em;
    font-size: clamp(19px, 2vw, 26px);
    line-height: 1.45;
    color: var(--muted);
  }

  .meta {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 14px 24px;
    margin-top: 28px;
    font: 500 15px/1 'DM Mono', monospace;
    color: var(--muted);
  }

  .badge {
    display: inline-flex;
    align-items: center;
    gap: 9px;
    font: 600 15px/1 'Gabarito', sans-serif;
    color: var(--ink);
  }

  .badge i,
  .status-row i {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    background: var(--green);
    flex: none;
  }

  .body-grid {
    display: grid;
    grid-template-columns: minmax(0, 220px) minmax(0, 760px);
    gap: clamp(40px, 6vw, 96px);
    align-items: start;
    margin-top: clamp(64px, 8vw, 104px);
  }

  .index {
    position: sticky;
    top: calc(var(--nav-h) + 24px);
  }

  .index-title {
    margin: 0 0 18px;
    padding-top: 18px;
    border-top: 1px solid var(--line);
    font: 600 14px/1 'DM Mono', monospace;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    color: var(--muted);
  }

  .index ol {
    margin: 0;
    padding: 0;
    list-style: none;
    display: grid;
    gap: 12px;
  }

  .index a {
    font: 600 17px/1.35 'Gabarito', sans-serif;
    color: var(--muted);
    text-decoration: none;
  }

  .index a:hover {
    color: var(--blue);
  }

  article {
    padding-bottom: 44px;
    margin-bottom: 44px;
    border-bottom: 1px solid var(--line);
  }

  article:last-child {
    padding-bottom: 0;
    margin-bottom: 0;
    border-bottom: 0;
  }

  article h2 {
    margin: 0 0 18px;
    font-size: clamp(30px, 3vw, 40px);
    line-height: 1.08;
    letter-spacing: -0.025em;
    font-weight: 800;
  }

  article p {
    margin: 0;
    max-width: 46em;
    font-size: 19px;
    line-height: 1.6;
    color: var(--muted);
  }

  ul {
    margin: 26px 0 0;
    padding: 0;
    list-style: none;
    display: grid;
    gap: 14px;
  }

  li {
    position: relative;
    padding-left: 24px;
    font-size: 18px;
    line-height: 1.5;
    color: var(--ink);
  }

  li::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0.62em;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--yellow);
  }

  .status-list {
    margin-top: 28px;
    border-top: 1px solid var(--line);
  }

  .status-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 20px;
    padding: 16px 0;
    border-bottom: 1px solid var(--line);
    font: 600 18px/1 'Gabarito', sans-serif;
  }

  .status-row > span:first-child {
    display: inline-flex;
    align-items: center;
    gap: 12px;
  }

  .state {
    color: var(--muted);
    font-weight: 500;
  }

  .cta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 32px;
    margin-top: clamp(72px, 9vw, 120px);
    padding: 40px 44px;
    border-radius: 28px;
    background: var(--soft);
  }

  .cta h2 {
    margin: 0 0 10px;
    font-size: clamp(28px, 3vw, 38px);
    line-height: 1.1;
    letter-spacing: -0.025em;
    font-weight: 800;
  }

  .cta p {
    margin: 0;
    font-size: 18px;
    color: var(--muted);
  }

  .cta .btn {
    flex: none;
  }

  @media (max-width: 700px) {
    h1 {
      white-space: normal;
    }
  }

  @media (max-width: 900px) {
    .info {
      padding-top: 44px;
    }

    .body-grid {
      grid-template-columns: minmax(0, 1fr);
      gap: 40px;
      margin-top: 56px;
    }

    .index {
      position: static;
    }

    .index ol {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .cta {
      align-items: flex-start;
      flex-direction: column;
      padding: 32px 28px;
    }
  }
</style>
