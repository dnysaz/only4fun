<script>
  import { navigate } from './router.svelte.js';

  function goInstant(event) {
    event.preventDefault();
    navigate('/instant');
  }
</script>

<section class="band">
  <div class="wrap grid2">
    <div class="copy">
      <p class="eyebrow">Zero config · live in seconds</p>
      <h2>Push it. We'll do the rest.</h2>
      <p class="lede">
        No servers, no setup calls, no config files. Paste a public GitHub link or drop a ZIP — we
        detect your framework, build on AWS, switch on Cloudflare SSL and hand you a live URL.
      </p>
      <p class="tech">
        <span class="dot" aria-hidden="true"></span>
        Powered by AWS servers and a compiled Go backend — fast, lightweight and secure, with
        Cloudflare SSL on every deploy.
      </p>
      <a class="btn" href="/instant" onclick={goInstant}>Deploy for free →</a>
      <p class="note">Free forever. No credit card. No limits.</p>
    </div>

    <div class="log" role="img" aria-label="Example deploy log for a GitHub repository">
      <div class="tl"><span></span><span></span><span></span></div>
      <div class="l"><span class="d">only4fun</span> $ import github.com/yoursite/portfolio</div>
      <div class="l d">Detected Next.js · zero config needed</div>
      <div class="l d">Building on AWS</div>
      <div class="l"><span class="g">✓</span> SSL active via Cloudflare</div>
      <div class="l"><span class="g">✓</span> Build finished in <span class="y">14s</span></div>
      <div class="l">Live at <span class="b">https://yoursite.only4.fun</span></div>
    </div>
  </div>
</section>

<style>
  /* Red theme band: teks putih di merah gelap supaya tetap terbaca. */
  .band {
    background: #c5221f;
    color: #fff;
    padding: 100px 0;
  }

  .grid2 {
    display: grid;
    grid-template-columns: 1fr 1.15fr;
    gap: 60px;
    align-items: center;
  }

  .eyebrow {
    margin: 0 0 20px;
    font: 500 13px/1 'DM Mono', monospace;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: rgba(255, 255, 255, 0.8);
  }

  h2 {
    font-size: clamp(38px, 5.2vw, 68px);
    line-height: 1;
    letter-spacing: -0.04em;
    font-weight: 800;
    margin: 0 0 22px;
    text-wrap: balance;
  }

  .lede {
    color: rgba(255, 255, 255, 0.92);
    font-size: 22px;
    max-width: 26em;
    margin: 0;
  }

  .tech {
    display: flex;
    align-items: baseline;
    gap: 12px;
    margin: 20px 0 0;
    font-size: 18px;
    font-weight: 600;
    color: #fff;
  }

  .tech .dot {
    flex: 0 0 auto;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--yellow);
  }

  .band .btn {
    /* Pill putih tetap kontras di atas merah, seperti pill gelap di band kuning. */
    background: #fff;
    color: #202124;
    margin-top: 28px;
    font-size: 20px;
  }

  .band .btn:hover {
    background: var(--yellow);
    color: #202124;
  }

  .note {
    margin: 18px 0 0;
    font-size: 17px;
    color: rgba(255, 255, 255, 0.82);
  }

  /* Log putih, bukan terminal gelap, agar selaras dengan permukaan situs. */
  .log {
    background: #fff;
    color: #202124;
    border-radius: 24px;
    padding: 26px 28px 30px;
    font: 400 16px/1.9 'DM Mono', monospace;
    overflow-x: auto;
  }

  .log .tl {
    display: flex;
    gap: 8px;
    margin-bottom: 18px;
  }

  .log .tl span {
    width: 13px;
    height: 13px;
    border-radius: 50%;
    background: var(--line);
  }

  .log .tl span:nth-child(1) {
    background: var(--red);
  }
  .log .tl span:nth-child(2) {
    background: var(--yellow);
  }
  .log .tl span:nth-child(3) {
    background: var(--green);
  }

  .log .l {
    white-space: nowrap;
    opacity: 0;
    animation: in 0.01s forwards;
  }

  .log .d {
    color: #5f6368;
  }
  .log .g {
    color: #1e8e3e;
  }
  .log .y {
    color: #a86400;
    font-weight: 700;
  }
  .log .b {
    color: #1967d2;
    font-weight: 700;
  }

  .log .l:nth-child(2) {
    animation-delay: 0.4s;
  }
  .log .l:nth-child(3) {
    animation-delay: 1.1s;
  }
  .log .l:nth-child(4) {
    animation-delay: 1.6s;
  }
  .log .l:nth-child(5) {
    animation-delay: 2s;
  }
  .log .l:nth-child(6) {
    animation-delay: 2.4s;
  }
  .log .l:nth-child(7) {
    animation-delay: 3s;
  }

  @keyframes in {
    to {
      opacity: 1;
    }
  }

  @media (max-width: 900px) {
    .grid2 {
      grid-template-columns: 1fr;
      gap: 36px;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .log .l {
      animation: none;
      opacity: 1;
    }
  }
</style>
