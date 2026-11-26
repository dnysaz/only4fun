<script>
  import Dots from './Dots.svelte';
  import { getRoute, navigate } from './router.svelte.js';

  const groups = [
    {
      title: 'Explore',
      links: [
        { label: 'Home', href: '/' },
        { label: 'How it works', href: '#how' },
        { label: 'Features', href: '#features' },
        { label: 'Instant Launch', href: '/instant' },
        { label: 'Launchpad', href: '/launchpad' },
        { label: 'Deploy new', href: '/deploy' },
      ],
    },
    {
      title: 'Resources',
      links: [
        { label: 'Docs', href: '/docs' },
        { label: 'Status', href: '/status' },
        { label: 'Support', href: '/support' },
      ],
    },
    {
      title: 'Legal',
      links: [
        { label: 'Terms', href: '/terms' },
        { label: 'Privacy', href: '/privacy' },
      ],
    },
  ];

  function go(event, href) {
    event.preventDefault();

    if (href.startsWith('#')) {
      const id = href.slice(1);
      const scrollToSection = () => {
        document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' });
        history.replaceState({}, '', `#${id}`);
      };

      if (getRoute() === '/') scrollToSection();
      else {
        navigate('/');
        requestAnimationFrame(() => requestAnimationFrame(scrollToSection));
      }
      return;
    }

    navigate(href);
  }
</script>

<footer>
  <div class="wrap inner">
    <div class="intro">
      <div class="brandline"><Dots size={26} /><span class="word">Only4<span class="fun">fun</span></span></div>
      <p>Push, build, live. Free hosting for the fun stuff — no config, no limits, no credit card.</p>
    </div>

    <nav class="groups" aria-label="Footer links">
      {#each groups as group}
        <div class="group">
          <h3>{group.title}</h3>
          <ul>
            {#each group.links as link}
              <li>
                <a href={link.href} onclick={(event) => go(event, link.href)}>{link.label}</a>
              </li>
            {/each}
          </ul>
        </div>
      {/each}
    </nav>
  </div>

  <div class="watermark-clip" aria-hidden="true">
    <div class="watermark">Only4<span class="fun">fun</span></div>
  </div>
</footer>

<style>
  footer {
    --footer-bg: #202124;
    --footer-ink: #f1f3f4;
    --footer-muted: #9aa0a6;
    background: var(--footer-bg);
    color: var(--footer-ink);
    padding: 88px 0 0;
    overflow: hidden;
  }

  .inner {
    display: grid;
    grid-template-columns: minmax(0, 1.2fr) minmax(0, 2fr);
    gap: 72px;
    padding-bottom: 72px;
  }

  .brandline {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 30px;
    font-weight: 800;
    letter-spacing: -0.045em;
    white-space: nowrap;
  }

  .brandline .word {
    display: inline-block;
  }

  .fun {
    color: var(--blue);
  }

  .intro p {
    max-width: 34ch;
    margin: 20px 0 0;
    font-size: 18px;
    line-height: 1.5;
    color: var(--footer-muted);
  }

  .groups {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 32px;
  }

  .group h3 {
    margin: 0 0 20px;
    font: 500 13px/1 'DM Mono', monospace;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: rgba(255, 255, 255, 0.45);
  }

  .group ul {
    display: grid;
    gap: 12px;
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .group a {
    font-size: 16px;
    text-decoration: none;
    color: var(--footer-ink);
  }

  .group a:hover {
    color: var(--yellow);
  }

  /* Bayang-bayang tulisan besar di paling bawah footer. */
  .watermark-clip {
    display: flex;
    align-items: flex-end;
    justify-content: center;
    height: clamp(150px, 24vw, 340px);
    overflow: hidden;
  }

  .watermark {
    font-size: clamp(120px, 24vw, 380px);
    font-weight: 800;
    line-height: 0.9;
    letter-spacing: -0.06em;
    white-space: nowrap;
    color: rgba(255, 255, 255, 0.055);
    transform: translateY(4%);
    user-select: none;
  }

  .watermark .fun {
    color: rgba(66, 133, 244, 0.38);
  }

  @media (max-width: 900px) {
    .inner {
      grid-template-columns: 1fr;
      gap: 48px;
      padding-bottom: 56px;
    }

    .groups {
      grid-template-columns: repeat(2, minmax(0, 1fr));
      row-gap: 40px;
    }
  }

  @media (max-width: 560px) {
    footer {
      padding-top: 64px;
    }

    .groups {
      grid-template-columns: 1fr;
    }

    .watermark-clip {
      height: 130px;
    }

    .watermark {
      font-size: 120px;
    }
  }
</style>
