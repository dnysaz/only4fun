package server

import (
	"fmt"
	"html"
	"strings"
)

// errorPage renders a branded error page that mirrors the public homepage:
// the same pill navbar, oversized headline, animated four-dot mark and footer.
func errorPage(statusCode int, title, highlight, message, ctaText, baseURL string) string {
	base := html.EscapeString(strings.TrimRight(baseURL, "/"))
	if base == "" {
		base = "https://only4.fun"
	}
	safeTitle := html.EscapeString(title)
	safeHighlight := html.EscapeString(highlight)
	safeCTA := html.EscapeString(ctaText)
	status := fmt.Sprintf("%d", statusCode)

	return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover">
  <meta name="color-scheme" content="light dark">
  <title>` + safeTitle + ` ` + safeHighlight + ` — only4.fun</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Gabarito:wght@400;500;600;700;800&family=DM+Mono:wght@400;500&display=swap" rel="stylesheet">
  <style>
    :root {
      --bg: #ffffff;
      --soft: #f1f3f4;
      --ink: #202124;
      --muted: #5f6368;
      --line: #dadce0;
      --blue: #4285f4;
      --red: #ea4335;
      --yellow: #fbbc04;
      --green: #34a853;
      --term: #202124;
      --term-ink: #e8eaed;
      --on-yellow: #202124;
      --nav-h: 92px;
      --page-x: clamp(28px, 4vw, 72px);
      color-scheme: light;
    }
    @media (prefers-color-scheme: dark) {
      :root:not([data-theme='light']) {
        --bg: #202124;
        --soft: #2d2e31;
        --ink: #e8eaed;
        --muted: #9aa0a6;
        --line: #3c4043;
        --term: #171717;
        color-scheme: dark;
      }
    }
    :root[data-theme='dark'] {
      --bg: #202124;
      --soft: #2d2e31;
      --ink: #e8eaed;
      --muted: #9aa0a6;
      --line: #3c4043;
      --term: #171717;
      color-scheme: dark;
    }
    html {
      scroll-behavior: smooth;
      scroll-padding-top: calc(var(--nav-h) + env(safe-area-inset-top, 0px));
      padding-bottom: env(safe-area-inset-bottom, 0px);
    }
    *, *::before, *::after { box-sizing: border-box; }
    body {
      margin: 0;
      padding-top: calc(var(--nav-h) + env(safe-area-inset-top, 0px));
      background-color: var(--bg);
      color: var(--ink);
      font-family: 'Gabarito', system-ui, sans-serif;
      font-size: 20px;
      line-height: 1.5;
      -webkit-font-smoothing: antialiased;
    }
    a { color: inherit; }
    :focus-visible { outline: 3px solid var(--blue); outline-offset: 3px; }
    .wrap { width: 100%; padding: 0 var(--page-x); }

    .btn {
      display: inline-block;
      background: var(--blue);
      color: #fff;
      padding: 15px 30px;
      border-radius: 999px;
      font: 600 19px/1 'Gabarito', sans-serif;
      text-decoration: none;
      border: 0;
      cursor: pointer;
      transition: background 0.15s;
    }
    .btn:hover { background: #2f72e0; }
    body .btn:hover { background: var(--blue); border-color: var(--blue); color: #fff; }

    /* Four-dot logo mark, shared by the nav and the footer. */
    .dots4 {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 3px;
      width: var(--dot-size, 26px);
      height: var(--dot-size, 26px);
    }
    .dots4 i { border-radius: 50%; }
    .dots4 i:nth-child(1) { background: var(--blue); }
    .dots4 i:nth-child(2) { background: var(--red); }
    .dots4 i:nth-child(3) { background: var(--yellow); }
    .dots4 i:nth-child(4) { background: var(--green); }

    nav {
      position: fixed;
      top: calc(env(safe-area-inset-top, 0px) + 12px);
      left: var(--page-x);
      right: var(--page-x);
      z-index: 50;
      display: flex;
      align-items: center;
      justify-content: space-between;
      height: calc(var(--nav-h) - 24px);
      padding: 0 28px;
      background-color: var(--bg);
      border: 1px solid transparent;
      border-radius: 999px;
      transition: border-color 160ms ease;
    }
    nav.scrolled { border-color: var(--line); }
    .logo {
      display: flex;
      align-items: center;
      gap: 10px;
      font-weight: 800;
      font-size: 28px;
      letter-spacing: -0.045em;
      white-space: nowrap;
      text-decoration: none;
    }
    .logo .word { display: inline-block; }
    .logo .fun { color: var(--blue); }
    nav ul {
      display: flex;
      gap: 32px;
      list-style: none;
      margin: 0;
      padding: 0;
      align-items: center;
    }
    nav a { text-decoration: none; font-weight: 500; color: var(--muted); }
    nav a:hover { color: var(--ink); }
    .nav-arrow {
      display: inline-flex;
      align-items: center;
      margin-left: 5px;
      transition: transform 160ms ease;
    }
    .nav-arrow svg { width: 15px; height: 15px; }
    nav a:hover .nav-arrow { transform: translate(2px, -2px); }
    .actions { display: flex; align-items: center; gap: 12px; }
    nav a.btn {
      color: #fff;
      padding: 12px 24px;
      font-size: 17px;
      white-space: nowrap;
    }

    main {
      min-height: calc(100vh - var(--nav-h));
      display: flex;
      flex-direction: column;
    }
    .hero {
      position: relative;
      padding: 64px 0 72px;
      isolation: isolate;
      overflow-x: clip;
    }
    .hero-inner { position: relative; }
    .copy { max-width: none; position: relative; z-index: 1; }

    /* Empat dot logo di sisi kanan yang kosong, bounce bergiliran. */
    .hero-dots {
      position: absolute;
      top: 215px;
      right: calc(var(--page-x) + 88px);
      z-index: 0;
      pointer-events: none;
    }
    .hero-dots .dots4 {
      grid-template-columns: repeat(4, 72px);
      grid-template-rows: 72px;
      width: auto;
      height: 72px;
      gap: 10px;
      transform-origin: center;
      animation: cluster-spin 5.6s ease-in-out infinite;
    }
    .hero-dots .dots4 i { animation: dot-cycle 5.6s ease-in-out infinite; }
    .hero-dots .dots4 i:nth-child(1),
    .hero-dots .dots4 i:nth-child(2) { --fx: 82px; --fy: -41px; }
    .hero-dots .dots4 i:nth-child(3),
    .hero-dots .dots4 i:nth-child(4) { --fx: -82px; --fy: 41px; }
    .hero-dots .dots4 i:nth-child(1) { animation-delay: 0s; }
    .hero-dots .dots4 i:nth-child(2) { animation-delay: 0.08s; }
    .hero-dots .dots4 i:nth-child(3) { animation-delay: 0.16s; }
    .hero-dots .dots4 i:nth-child(4) { animation-delay: 0.24s; }
    @keyframes dot-cycle {
      0% { transform: translate(0, 0); }
      12% { transform: translateY(-34px); }
      24% { transform: translate(0, 0); }
      36% { transform: translateY(-34px); }
      46% { transform: translate(0, 0); }
      56%, 78% { transform: translate(var(--fx, 0px), var(--fy, 0px)) scale(1); }
      83% { transform: translate(var(--fx, 0px), var(--fy, 0px)) scale(1.28); }
      88% { transform: translate(var(--fx, 0px), var(--fy, 0px)) scale(0.72); }
      93% { transform: translate(0, 0) scale(1); }
      97% { transform: translateY(-24px) scale(1); }
      100% { transform: translate(0, 0); }
    }
    @keyframes cluster-spin {
      0%, 56% { transform: rotate(0deg); }
      78% { transform: rotate(360deg); }
      78.01%, 100% { transform: rotate(0deg); }
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
    .pulse { width: 8px; height: 8px; border-radius: 50%; background: var(--green); }
    h1 {
      font-size: clamp(52px, 8vw, 112px);
      line-height: 0.93;
      letter-spacing: -0.045em;
      font-weight: 800;
      margin: 0;
      text-wrap: balance;
    }
    .hl {
      background: linear-gradient(transparent 62%, rgba(251, 188, 4, 0.5) 62%);
      padding: 0 0.04em;
    }
    .lower {
      display: grid;
      grid-template-columns: minmax(0, 760px);
      align-items: start;
      margin-top: 22px;
    }
    .intro { min-width: 0; padding-top: 26px; }
    .lead {
      font-size: clamp(19px, 1.7vw, 24px);
      line-height: 1.45;
      max-width: 30em;
      margin: 0;
      color: var(--muted);
    }
    .lead strong {
      color: var(--ink);
      text-decoration: underline;
      text-decoration-thickness: 2px;
      text-underline-offset: 0.18em;
    }
    .hero-actions {
      display: flex;
      align-items: center;
      flex-wrap: wrap;
      gap: 18px;
      margin-top: 32px;
    }
    .instant-link {
      display: inline-block;
      font-size: 16px;
      color: var(--muted);
      text-decoration: underline;
      text-decoration-thickness: 1.5px;
      text-underline-offset: 0.2em;
    }
    .instant-link:hover { color: var(--blue); }

    footer {
      --footer-bg: #202124;
      --footer-ink: #f1f3f4;
      --footer-muted: #9aa0a6;
      margin-top: auto;
      background: var(--footer-bg);
      color: var(--footer-ink);
      padding: 88px 0 0;
      overflow: hidden;
    }
    footer .inner {
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
    .brandline .word { display: inline-block; }
    .brandline .fun { color: var(--blue); }
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
    .group a { font-size: 16px; color: var(--footer-ink); text-decoration: none; }
    .group a:hover { color: var(--yellow); }
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
      user-select: none;
      color: rgba(255, 255, 255, 0.055);
      transform: translateY(4%);
    }
    .watermark .fun { color: rgba(66, 133, 244, 0.38); }

    @media (max-width: 1050px) { .hero-dots { display: none; } }
    @media (max-width: 900px) {
      body { font-size: 18px; }
      nav ul li:not(.actions) { display: none; }
      .hero { padding: 26px 0 72px; }
      .eyebrow { margin-bottom: 20px; font-size: 12px; }
      .lower { margin-top: 24px; }
      footer .inner {
        grid-template-columns: 1fr;
        gap: 48px;
        padding-bottom: 56px;
      }
      .groups {
        grid-template-columns: repeat(2, minmax(0, 1fr));
        row-gap: 40px;
      }
    }
    @media (max-width: 700px) {
      :root { --page-x: 18px; }
      nav { padding: 0 18px; }
      .logo { font-size: 24px; gap: 8px; }
      nav a.btn { padding: 10px 18px; font-size: 15px; }
      .groups { grid-template-columns: 1fr; }
    }
    @media (max-width: 560px) {
      footer { padding-top: 64px; }
      .watermark-clip { height: 130px; }
      .watermark { font-size: 120px; }
    }
    @media (prefers-reduced-motion: reduce) {
      html { scroll-behavior: auto; }
      .btn { transition: none; }
      .hero-dots .dots4,
      .hero-dots .dots4 i { animation: none; }
    }
  </style>
</head>
<body>
  <nav>
    <a class="logo" href="` + base + `/"><span class="dots4" style="--dot-size:26px" aria-hidden="true"><i></i><i></i><i></i><i></i></span><span class="word">Only4<span class="fun">fun</span></span></a>
    <ul>
      <li><a href="` + base + `/#how">How it works</a></li>
      <li><a href="` + base + `/#features">Features</a></li>
      <li><a href="` + base + `/instant">Instant Launch <span class="nav-arrow" aria-hidden="true"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M7 17 17 7" /><path d="M8 7h9v9" /></svg></span></a></li>
      <li class="actions"><a class="btn" href="` + base + `/launchpad">Get started</a></li>
    </ul>
  </nav>

  <main>
    <header class="hero">
      <div class="wrap hero-inner">
        <div class="copy">
          <p class="eyebrow"><span class="pulse" aria-hidden="true"></span>Error ` + status + `</p>
          <h1><span>` + safeTitle + `</span><br /><span class="hl">` + safeHighlight + `</span></h1>
          <div class="lower">
            <div class="intro">
              <p class="lead">` + message + `</p>
              <div class="hero-actions">
                <a class="btn" href="` + base + `/deploy">` + safeCTA + `</a>
                <a class="instant-link" href="` + base + `/">Go to homepage</a>
              </div>
            </div>
          </div>
        </div>
        <div class="hero-dots" aria-hidden="true"><span class="dots4" style="--dot-size:150px"><i></i><i></i><i></i><i></i></span></div>
      </div>
    </header>
  </main>

  <footer>
    <div class="wrap inner">
      <div class="intro">
        <div class="brandline"><span class="dots4" style="--dot-size:26px" aria-hidden="true"><i></i><i></i><i></i><i></i></span><span class="word">Only4<span class="fun">fun</span></span></div>
        <p>Push, build, live. Free hosting for the fun stuff — no config, no limits, no credit card.</p>
      </div>
      <nav class="groups" aria-label="Footer links">
        <div class="group">
          <h3>Explore</h3>
          <ul>
            <li><a href="` + base + `/">Home</a></li>
            <li><a href="` + base + `/#how">How it works</a></li>
            <li><a href="` + base + `/#features">Features</a></li>
            <li><a href="` + base + `/instant">Instant Launch</a></li>
            <li><a href="` + base + `/launchpad">Launchpad</a></li>
            <li><a href="` + base + `/deploy">Deploy new</a></li>
          </ul>
        </div>
        <div class="group">
          <h3>Resources</h3>
          <ul>
            <li><a href="` + base + `/docs">Docs</a></li>
            <li><a href="` + base + `/status">Status</a></li>
            <li><a href="` + base + `/support">Support</a></li>
          </ul>
        </div>
        <div class="group">
          <h3>Legal</h3>
          <ul>
            <li><a href="` + base + `/terms">Terms</a></li>
            <li><a href="` + base + `/privacy">Privacy</a></li>
          </ul>
        </div>
      </nav>
    </div>
    <div class="watermark-clip" aria-hidden="true"><div class="watermark">Only4<span class="fun">fun</span></div></div>
  </footer>

  <script>
    (function () {
      var nav = document.querySelector('body > nav');
      if (!nav) return;
      function sync() {
        if (window.scrollY > 8) nav.classList.add('scrolled');
        else nav.classList.remove('scrolled');
      }
      sync();
      window.addEventListener('scroll', sync, { passive: true });
    })();
  </script>
</body>
</html>`
}
