<div class="art" aria-hidden="true">
  <svg viewBox="0 0 360 280" fill="none" xmlns="http://www.w3.org/2000/svg">
    <!-- dashed flight path -->
    <path
      d="M232 86c26-22 52-34 74-40"
      stroke="var(--muted)"
      stroke-opacity="0.45"
      stroke-width="2"
      stroke-dasharray="5 7"
      stroke-linecap="round"
    />

    <!-- browser window -->
    <rect
      x="38"
      y="58"
      width="252"
      height="174"
      rx="18"
      fill="var(--bg)"
      stroke="var(--line)"
      stroke-width="2"
    />
    <path d="M38 94h252" stroke="var(--line)" stroke-width="2" />
    <rect x="56" y="70" width="150" height="16" rx="8" fill="var(--soft)" />
    <circle cx="68" cy="78" r="3" fill="var(--muted)" opacity="0.7" />

    <!-- page content skeleton -->
    <rect x="58" y="112" width="126" height="13" rx="6.5" fill="var(--ink)" opacity="0.78" />
    <rect x="58" y="136" width="176" height="8" rx="4" fill="var(--muted)" opacity="0.35" />
    <rect x="58" y="152" width="140" height="8" rx="4" fill="var(--muted)" opacity="0.35" />
    <rect x="58" y="176" width="76" height="20" rx="10" fill="var(--blue)" />
    <rect x="58" y="208" width="196" height="8" rx="4" fill="var(--soft)" />

    <!-- success badge -->
    <circle cx="268" cy="206" r="27" fill="var(--green)" />
    <path
      d="M256 206l8.5 9.5L281 196"
      stroke="#fff"
      stroke-width="5"
      stroke-linecap="round"
      stroke-linejoin="round"
    />

    <!-- paper plane -->
    <g class="plane">
      <path d="M300 22l44 22-44 10 6-16z" fill="var(--blue)" />
      <path d="M300 22l30 24-24 8 6-16z" fill="var(--blue)" opacity="0.55" />
    </g>

    <!-- confetti -->
    <circle cx="318" cy="112" r="7" fill="var(--yellow)" />
    <rect
      x="300"
      y="245"
      width="15"
      height="15"
      rx="3"
      fill="var(--red)"
      transform="rotate(-16 307 252)"
    />
    <path d="M22 132l13 24H9z" fill="var(--yellow)" />
    <circle cx="30" cy="252" r="6" fill="var(--green)" />
    <path d="M320 176l5 11 11 5-11 5-5 11-5-11-11-5 11-5z" fill="var(--red)" opacity="0.75" />
  </svg>
</div>

<style>
  .art {
    width: 420px;
    max-width: 100%;
    justify-self: end;
    animation: float 6s ease-in-out infinite alternate;
  }

  svg {
    display: block;
    width: 100%;
    height: auto;
    overflow: visible;
  }

  @keyframes float {
    from {
      transform: translateY(4px);
    }
    to {
      transform: translateY(-8px);
    }
  }

  @media (max-width: 900px) {
    .art {
      width: 100%;
      max-width: 300px;
      justify-self: start;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .art {
      animation: none;
    }
  }
</style>
