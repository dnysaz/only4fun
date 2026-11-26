<script>
  import Icon from './Icon.svelte';
  import FrameworkMarquee from './FrameworkMarquee.svelte';

  const features = [
    {
      icon: 'world',
      tone: 'blue',
      title: 'Your own subdomain',
      text: 'Every project gets its own',
      code: 'yoursite.only4.fun',
      tail: 'address, yours to keep.',
    },
    {
      icon: 'shield',
      tone: 'red',
      title: 'SSL by Cloudflare',
      text: 'Every site is served over HTTPS behind Cloudflare, with DDoS protection built in.',
    },
    {
      icon: 'branch',
      tone: 'yellow',
      title: 'Auto-deploy on push',
      text: 'Connect GitHub and every push to your main branch updates the live site.',
    },
    {
      icon: 'server',
      tone: 'green',
      title: 'Singapore servers',
      text: 'Builds and hosting run in Singapore, close to visitors across Southeast Asia.',
    },
    {
      icon: 'bolt',
      tone: 'blue',
      title: 'Fast builds',
      text: 'Dependencies are cached, so most projects build and go live in seconds.',
    },
    {
      icon: 'infinity',
      tone: 'red',
      title: 'Free, no limits',
      text: 'No bandwidth caps, no project limits and no credit card. Ever.',
    },
  ];

</script>

<section class="feat" id="features">
  <div class="wrap">
    <h2>Works with every frontend framework.</h2>
    <FrameworkMarquee />
    <div class="fl">
      {#each features as feature}
        <div>
          <div class="ic {feature.tone}"><Icon name={feature.icon} /></div>
          <h3>{feature.title}</h3>
          <p>
            {feature.text}
            {#if feature.code}
              <code>{feature.code}</code>
              {feature.tail}
            {/if}
          </p>
        </div>
      {/each}
    </div>
  </div>
</section>

<style>
  .feat {
    padding: 28px 0 140px;
  }

  h2 {
    font-size: clamp(38px, 5.2vw, 68px);
    line-height: 1;
    letter-spacing: -0.04em;
    font-weight: 800;
    margin: 0 0 34px;
  }

  .fl {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 72px 44px;
    margin-top: 88px;
  }

  .ic {
    display: grid;
    place-items: center;
    width: 48px;
    height: 48px;
    margin-bottom: 26px;
    border-radius: 50%;
    color: #fff;
  }

  .ic.blue {
    background: var(--blue);
  }
  .ic.red {
    background: var(--red);
  }
  .ic.yellow {
    background: var(--yellow);
    /* dark glyph keeps contrast on yellow */
    color: #202124;
  }
  .ic.green {
    background: var(--green);
  }

  .fl h3 {
    margin: 0 0 12px;
    font-size: 26px;
    letter-spacing: -0.02em;
    font-weight: 700;
  }

  .fl p {
    margin: 0;
    color: var(--muted);
    font-size: 19px;
    line-height: 1.55;
  }

  .fl code {
    font: 500 0.78em 'DM Mono', monospace;
    background: var(--soft);
    padding: 2px 8px;
    border-radius: 8px;
  }

  @media (max-width: 900px) {
    .fl {
      grid-template-columns: 1fr 1fr;
      gap: 60px 36px;
      margin-top: 72px;
    }
  }

  @media (max-width: 560px) {
    .feat {
      padding: 16px 0 104px;
    }

    .fl {
      grid-template-columns: 1fr;
      gap: 56px;
      margin-top: 64px;
    }
  }
</style>
