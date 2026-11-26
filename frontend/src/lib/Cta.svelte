<script>
  import { navigate } from './router.svelte.js';

  function goInstant(event) {
    event.preventDefault();
    navigate('/instant');
  }
</script>

<section class="cta">
  <div class="wrap">
    <h2>Deploy your fun projects — only on Only4fun.</h2>
    <div class="actions">
      <a class="btn" href="#claim">Start Deploy</a>
      <a class="quick" href="/instant" onclick={goInstant}>Deploy instantly, no sign in →</a>
    </div>
    <p>Free with no limits. No credit card.</p>
  </div>
</section>

<style>
  .cta {
    background: var(--yellow);
    color: var(--on-yellow);
    padding: 110px 0;
  }

  h2 {
    font-size: clamp(46px, 7.4vw, 106px);
    line-height: 1.02;
    max-width: 26ch;
    margin: 0 0 40px;
    letter-spacing: -0.04em;
    font-weight: 800;
  }

  .actions {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 24px;
  }

  .cta .btn {
    /* fixed dark pill: the band is yellow in both themes */
    background: #202124;
    color: #fff;
    font-size: 21px;
    padding: 20px 38px;
  }

  .cta .btn:hover {
    background: #000;
  }

  .quick {
    font-size: 19px;
    font-weight: 600;
    text-decoration: underline;
    text-decoration-thickness: 2px;
    text-underline-offset: 0.2em;
  }

  .quick:hover {
    color: var(--blue);
  }

  .cta p {
    margin: 24px 0 0;
    font-size: 19px;
  }
</style>
