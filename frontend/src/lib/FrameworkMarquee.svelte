<script>
  import { frameworks } from './frameworks.ts';
  import FrameworkIcon from './FrameworkIcon.svelte';
</script>

<div class="marquee" aria-label="Supported frontend frameworks">
  <div class="track">
    {#each [0, 1] as copy (copy)}
      <div class="group" aria-hidden={copy === 1}>
        {#each frameworks as framework (framework.id)}
          <span class="item" title={framework.label}>
            <FrameworkIcon {framework} size={76} />
          </span>
        {/each}
      </div>
    {/each}
  </div>
</div>

<style>
  .marquee {
    /* full bleed keluar dari padding .wrap, tanpa menambah scroll horizontal */
    width: 100vw;
    max-width: none;
    margin-left: calc(50% - 50vw);
    overflow: hidden;
    padding: 34px 0;
  }

  .track {
    display: flex;
    width: max-content;
    animation: slide 40s linear infinite;
    will-change: transform;
  }

  .marquee:hover .track {
    animation-play-state: paused;
  }

  .group {
    display: flex;
    align-items: center;
    gap: 48px;
    /* jarak yang sama di ujung supaya -50% benar-benar mulus */
    padding-right: 48px;
  }

  .item {
    display: grid;
    place-items: center;
    flex: 0 0 auto;
    width: 76px;
    height: 76px;
    background: transparent;
    border: 0;
    border-radius: 0;
  }

  @keyframes slide {
    to {
      transform: translateX(-50%);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .track {
      animation: none;
    }
  }
</style>
