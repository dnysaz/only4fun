<script>
  import { onMount } from 'svelte';
  import Dots from './Dots.svelte';
  import { api } from './api.js';
  import { Loading } from './loading.ts';
  import { session } from './session.svelte.js';
  import { navigate } from './router.svelte.js';
  import { MAX_LENGTH, sanitize, validateName } from './subdomain.js';

  let name = $state('');
  let claimed = $state(false);
  let inputEl = $state(null);

  function evaluate(value) {
    const check = validateName(value);
    if (check.ok) return { text: `${value}.only4.fun is available.`, tone: 'ok', ok: true };
    if (check.reason === 'empty') {
      return { text: "Type a name to check if it's available.", tone: '', ok: false };
    }
    if (check.reason === 'short') {
      return { text: 'Use at least 3 characters.', tone: 'bad', ok: false };
    }
    if (check.reason === 'hyphen') {
      return { text: "Names can't start or end with a hyphen.", tone: 'bad', ok: false };
    }
    if (check.reason === 'blocked') {
      return { text: `"${value}" is not allowed. Try another name.`, tone: 'bad', ok: false };
    }
    return { text: `${value}.only4.fun is reserved. Try another name.`, tone: 'bad', ok: false };
  }

  let local = $derived(evaluate(name));

  // Hasil cek dari backend (butuh sesi login).
  let server = $state({ name: '', available: null, error: '', checking: false });

  $effect(() => {
    const value = local.ok ? name : '';
    if (!session.user || !value) {
      server = { name: '', available: null, error: '', checking: false };
      return;
    }

    let cancelled = false;
    const timer = setTimeout(async () => {
      server = { name: value, available: null, error: '', checking: true };
      try {
        const data = await api.checkSubdomain(value);
        if (cancelled) return;
        server = {
          name: value,
          available: !!data?.available,
          error: data?.available ? '' : (data?.error ?? 'this name is not available'),
          checking: false,
        };
      } catch (err) {
        if (cancelled) return;
        server = { name: value, available: null, error: err?.message ?? 'check failed', checking: false };
      }
    }, 400);

    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  });

  let status = $derived.by(() => {
    if (claimed && local.ok) {
      return {
        text: `${name}.only4.fun is ready. Finish the setup in your launchpad.`,
        tone: 'ok',
      };
    }
    if (!name) return { text: local.text, tone: '' };
    if (!local.ok) return { text: local.text, tone: local.tone };
    if (!session.user) {
      return { text: `${name}.only4.fun looks good. Sign in to claim it.`, tone: 'ok' };
    }
    if (server.checking) return { text: '', tone: '', checking: true };
    if (server.available === false) return { text: server.error, tone: 'bad' };
    if (server.error) return { text: server.error, tone: 'bad' };
    if (server.available === true) return { text: `${name}.only4.fun is available.`, tone: 'ok' };
    return { text: `${name}.only4.fun looks good.`, tone: '' };
  });

  function onInput(event) {
    name = sanitize(event.currentTarget.value);
    claimed = false;
  }

  function onSubmit(event) {
    event.preventDefault();
    if (!name) {
      inputEl?.focus();
      return;
    }
    if (!local.ok || server.available === false) return;

    claimed = true;
    const next = `/?claim=${encodeURIComponent(name)}`;
    if (session.user) navigate(`/deploy?claim=${encodeURIComponent(name)}`);
    else window.location.href = api.loginUrl(next);
  }

  function goInstant(event) {
    event.preventDefault();
    navigate('/instant');
  }

  function goHome(event) {
    event.preventDefault();
    navigate('/');
  }

  onMount(() => {
    const claim = new URLSearchParams(window.location.search).get('claim');
    if (claim) name = sanitize(claim);
  });
</script>

<header class="hero" id="claim">
  <div class="wrap hero-inner">
    <div class="copy">
      <p class="eyebrow"><span class="pulse" aria-hidden="true"></span>Free forever · no credit card</p>
      <h1>
        <a class="home-title" href="/" onclick={goHome}
          >Ship any static sites<br /><span class="hl">in seconds.</span></a
        >
      </h1>

      <div class="lower">
        <div class="intro">
          <p class="lead">
            Upload a ZIP, paste a GitHub link, or sign in with GitHub. Your site goes live at <span
              class="domain">yoursite.only4.fun</span
            >.
          </p>
          <div class="claimbox">
            <form class="bar" onsubmit={onSubmit} aria-label="Check subdomain">
              <span class="field">
                <input
                  bind:this={inputEl}
                  value={name}
                  oninput={onInput}
                  type="text"
                  placeholder="yoursite"
                  maxlength={MAX_LENGTH}
                  autocomplete="off"
                  spellcheck="false"
                  aria-label="Subdomain name"
                />
                <span class="suffix">.only4.fun</span>
              </span>
              <button class="btn" type="submit">
                Start deploy
                <svg
                  viewBox="0 0 24 24"
                  width="17"
                  height="17"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2.2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  aria-hidden="true"
                >
                  <path d="M7 17L17 7" />
                  <path d="M8 7h9v9" />
                </svg>
              </button>
            </form>
            <div class="status {status.tone}" aria-live="polite">
              {#if status.checking}
                <Loading label="Checking availability…" size={12} inline />
              {:else}
                {status.text}
              {/if}
            </div>
            <a class="instant-link" href="/instant" onclick={goInstant}
              >No account? Deploy a static site instantly →</a
            >
          </div>
        </div>
      </div>
    </div>
    <div class="hero-dots" aria-hidden="true"><Dots size={150} /></div>
  </div>
</header>

<style>
  .hero {
    position: relative;
    padding: 64px 0 72px;
    isolation: isolate;
    overflow-x: clip;
  }

  .hero-inner {
    position: relative;
  }

  /* Empat dot logo di sisi kanan yang kosong, bounce bergiliran. */
  .hero-dots {
    position: absolute;
    top: 215px;
    right: calc(var(--page-x) + 88px);
    z-index: 0;
    pointer-events: none;
  }

  .hero-dots :global(.dots4) {
    grid-template-columns: repeat(4, 72px);
    grid-template-rows: 72px;
    width: auto;
    height: 72px;
    gap: 10px;
    transform-origin: center;
    animation: cluster-spin 5.6s ease-in-out infinite;
  }

  .hero-dots :global(.dots4 i) {
    animation: dot-cycle 5.6s ease-in-out infinite;
  }

  /* Dua dot atas dan dua dot bawah bertemu di tengah jadi formasi 2x2. */
  .hero-dots :global(.dots4 i:nth-child(1)),
  .hero-dots :global(.dots4 i:nth-child(2)) {
    --fx: 82px;
    --fy: -41px;
  }

  .hero-dots :global(.dots4 i:nth-child(3)),
  .hero-dots :global(.dots4 i:nth-child(4)) {
    --fx: -82px;
    --fy: 41px;
  }

  .hero-dots :global(.dots4 i:nth-child(1)) {
    animation-delay: 0s;
  }

  .hero-dots :global(.dots4 i:nth-child(2)) {
    animation-delay: 0.08s;
  }

  .hero-dots :global(.dots4 i:nth-child(3)) {
    animation-delay: 0.16s;
  }

  .hero-dots :global(.dots4 i:nth-child(4)) {
    animation-delay: 0.24s;
  }

  @keyframes dot-cycle {
    0% {
      transform: translate(0, 0);
    }
    12% {
      transform: translateY(-34px);
    }
    24% {
      transform: translate(0, 0);
    }
    36% {
      transform: translateY(-34px);
    }
    46% {
      transform: translate(0, 0);
    }
    56%,
    78% {
      transform: translate(var(--fx, 0px), var(--fy, 0px)) scale(1);
    }
    83% {
      transform: translate(var(--fx, 0px), var(--fy, 0px)) scale(1.28);
    }
    88% {
      transform: translate(var(--fx, 0px), var(--fy, 0px)) scale(0.72);
    }
    93% {
      transform: translate(0, 0) scale(1);
    }
    97% {
      transform: translateY(-24px) scale(1);
    }
    100% {
      transform: translate(0, 0);
    }
  }

  @keyframes cluster-spin {
    0%,
    56% {
      transform: rotate(0deg);
    }
    78% {
      transform: rotate(360deg);
    }
    78.01%,
    100% {
      transform: rotate(0deg);
    }
  }

  .copy {
    max-width: none;
    position: relative;
    z-index: 1;
  }

  .eyebrow {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    margin: 0 0 26px;
    padding: 8px 16px 8px 14px;
    border: 1px solid var(--line);
    border-radius: 999px;
    font: 500 14px/1 'DM Mono', monospace;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--muted);
  }

  .pulse {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--green);
  }

  h1 {
    font-size: clamp(52px, 8vw, 112px);
    line-height: 0.93;
    letter-spacing: -0.045em;
    font-weight: 800;
    margin: 0;
    text-wrap: balance;
  }

  .home-title {
    display: block;
    width: fit-content;
    color: inherit;
    text-decoration: none;
    cursor: pointer;
  }

  .home-title:hover .hl {
    background: linear-gradient(transparent 58%, rgba(251, 188, 4, 0.85) 58%);
  }

  .hl {
    background: linear-gradient(transparent 62%, rgba(251, 188, 4, 0.5) 62%);
    padding: 0 0.04em;
  }

  /* Headline runs full width; the lead + form sit directly underneath. */
  .lower {
    display: grid;
    grid-template-columns: minmax(0, 760px);
    align-items: start;
    margin-top: 22px;
  }

  .intro {
    min-width: 0;
    /* small offset so the copy does not hug the headline */
    padding-top: 26px;
  }

  .lead {
    font-size: clamp(19px, 1.7vw, 24px);
    line-height: 1.45;
    max-width: 30em;
    margin: 0;
    color: var(--muted);
  }

  .domain {
    color: var(--ink);
    text-decoration: underline;
    text-decoration-thickness: 2px;
    text-underline-offset: 0.18em;
  }

  .claimbox {
    margin-top: 32px;
    max-width: 700px;
  }

  .bar {
    display: flex;
    align-items: center;
    background: var(--bg);
    border: 2px solid var(--line);
    border-radius: 999px;
    padding: 8px 8px 8px 26px;
    gap: 6px;
    transition: border-color 0.15s;
  }

  .bar:focus-within {
    border-color: var(--blue);
  }

  .field {
    display: flex;
    align-items: center;
    gap: 2px;
    flex: 1;
    min-width: 0;
  }

  .bar input {
    flex: 1;
    min-width: 0;
    border: 0;
    background: transparent;
    font: 500 clamp(19px, 2.2vw, 27px) / 1 'DM Mono', monospace;
    color: var(--ink);
    text-align: left;
    padding: 13px 0;
    outline: none;
    letter-spacing: -0.04em;
  }

  .bar input::placeholder {
    color: #bdc1c6;
  }

  .bar .suffix {
    font: 500 clamp(19px, 2.2vw, 27px) / 1 'DM Mono', monospace;
    letter-spacing: -0.04em;
    color: var(--muted);
    white-space: nowrap;
  }

  .bar .btn {
    display: inline-flex;
    align-items: center;
    gap: 9px;
    margin-left: 8px;
    white-space: nowrap;
    padding: 17px 26px;
    font-size: 18px;
  }

  .status {
    min-height: 30px;
    margin: 14px 0 0 26px;
    font-size: 17px;
    color: var(--muted);
  }

  .status.ok {
    color: var(--green);
  }

  .status.bad {
    color: var(--red);
  }

  .instant-link {
    display: inline-block;
    margin: 2px 0 0 26px;
    font-size: 16px;
    color: var(--muted);
    text-decoration: underline;
    text-decoration-thickness: 1.5px;
    text-underline-offset: 0.2em;
  }

  .instant-link:hover {
    color: var(--blue);
  }

  @media (max-width: 1050px) {
    .hero-dots {
      display: none;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .hero-dots :global(.dots4),
    .hero-dots :global(.dots4 i) {
      animation: none;
    }
  }

  @media (max-width: 900px) {
    .hero {
      padding: 26px 0 72px;
    }

    .eyebrow {
      margin-bottom: 20px;
      font-size: 12px;
    }

    .lower {
      margin-top: 24px;
    }

    .claimbox {
      margin-top: 26px;
    }

    /* mobile: a real input field on top, full-width button below */
    .bar {
      flex-direction: column;
      align-items: stretch;
      gap: 12px;
      border: 0;
      border-radius: 0;
      background: transparent;
      padding: 0;
    }

    .field {
      align-items: center;
      gap: 8px;
      width: 100%;
      border: 2px solid var(--line);
      border-radius: 16px;
      background: var(--bg);
      padding: 13px 16px;
      transition: border-color 0.15s;
    }

    .field:focus-within {
      border-color: var(--blue);
    }

    .bar input {
      width: 100%;
      flex: 1 1 auto;
      font-size: 18px;
      padding: 2px 0;
      letter-spacing: -0.02em;
    }

    .bar .suffix {
      flex: 0 0 auto;
      font-size: 16px;
      letter-spacing: -0.02em;
      color: var(--muted);
    }

    .bar .btn {
      margin-left: 0;
      width: 100%;
      justify-content: center;
      padding: 16px 20px;
      font-size: 17px;
    }

    .status {
      margin: 10px 2px 0;
      font-size: 15px;
      line-height: 1.35;
    }

    .instant-link {
      margin: 8px 2px 0;
      font-size: 15px;
    }
  }
</style>
