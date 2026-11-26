<script>
  import Illustration from './Illustration.svelte';
</script>

<section class="statement" id="how">
  <div class="wrap">
    <div class="panel">
      <div class="copy">
        <p class="tag">Fun, not a chore</p>
        <h2>Forget slow, complicated deploys.</h2>
        <p class="body">Use Only4fun for your fun projects. Deploying should be enjoyable, not a chore.</p>
      </div>
      <Illustration />

      <div class="steps">
        <p class="steps-label">How it works</p>
        <div class="steps-grid">
          <div class="step">
            <div class="num">01</div>
            <h3>Add your files</h3>
            <p>Paste a GitHub link or upload a ZIP. Static HTML, CSS and JS work right away.</p>
          </div>
          <div class="step">
            <div class="num">02</div>
            <h3>We deploy</h3>
            <p>Your files are built on our servers and published in seconds.</p>
          </div>
          <div class="step">
            <div class="num">03</div>
            <h3>SSL is on</h3>
            <p>Cloudflare gives every site HTTPS automatically. No certificates to manage.</p>
          </div>
          <div class="step">
            <div class="num">04</div>
            <h3>Launch</h3>
            <p>Your site is live at <code>yoursite.only4.fun</code>. Share the link and keep building.</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</section>

<style>
  /* Blue theme card: biru ada di dalam panel, bukan di background section. */
  .statement {
    padding: 96px 0;
  }

  .panel {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    column-gap: 56px;
    row-gap: 0;
    align-items: center;
    padding: 60px 64px;
    border-radius: 36px;
    /* Biru tema versi gelap supaya teks putih nyaman dibaca. */
    background: #1967d2;
    color: #fff;
  }

  /* Ilustrasi memakai aksen kuning + jalur putih di atas biru. */
  .panel :global(.art) {
    --blue: var(--yellow);
    --muted: #202124;
  }

  .panel :global(.art svg > path:first-child) {
    stroke: rgba(255, 255, 255, 0.75);
  }

  .copy {
    min-width: 0;
  }

  .tag {
    margin: 0 0 20px;
    font: 500 13px/1 'DM Mono', monospace;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: #fff;
  }

  h2 {
    font-size: clamp(32px, 4vw, 54px);
    line-height: 0.98;
    letter-spacing: -0.04em;
    font-weight: 800;
    margin: 0;
    text-wrap: balance;
  }

  .body {
    color: #fff;
    font-size: clamp(18px, 1.5vw, 21px);
    line-height: 1.5;
    margin: 20px 0 0;
  }

  /* How it works: empat langkah ringkas di bawah headline. */
  .steps {
    grid-column: 1 / -1;
    margin-top: 52px;
    padding-top: 30px;
    border-top: 1px solid rgba(255, 255, 255, 0.35);
  }

  .steps-label {
    margin: 0 0 24px;
    font: 500 13px/1 'DM Mono', monospace;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: #fff;
  }

  .steps-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .step {
    min-width: 0;
    padding: 0 26px;
  }

  .step:first-child {
    padding-left: 0;
  }

  .step + .step {
    border-left: 1px solid rgba(255, 255, 255, 0.35);
  }

  .num {
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    margin-bottom: 20px;
    border-radius: 50%;
    background: var(--yellow);
    color: var(--on-yellow);
    font: 700 15px/1 'DM Mono', monospace;
  }

  .step h3 {
    margin: 0 0 10px;
    font-size: 22px;
    line-height: 1.1;
    letter-spacing: -0.02em;
    font-weight: 700;
  }

  .step p {
    margin: 0;
    color: #fff;
    font-size: 17px;
    line-height: 1.5;
  }

  .step code {
    font: 500 0.88em 'DM Mono', monospace;
    color: #fff;
    background: rgba(255, 255, 255, 0.16);
    padding: 1px 5px;
    border-radius: 5px;
  }

  @media (max-width: 900px) {
    .statement {
      padding: 64px 0;
    }

    .panel {
      grid-template-columns: minmax(0, 1fr);
      gap: 36px;
      padding: 36px 28px 32px;
      border-radius: 28px;
    }

    h2 {
      max-width: none;
    }

    .steps {
      margin-top: 0;
    }

    .steps-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 28px 0;
    }

    .step {
      padding: 0 22px;
    }

    .step:nth-child(odd) {
      padding-left: 0;
      border-left: 0;
    }
  }

  @media (max-width: 560px) {
    .steps-grid {
      grid-template-columns: 1fr;
      gap: 22px;
    }

    .step,
    .step:nth-child(odd),
    .step:nth-child(even) {
      padding: 0;
      border-left: 0;
    }

    .step + .step {
      padding-top: 22px;
      border-top: 1px solid rgba(255, 255, 255, 0.35);
    }
  }
</style>
