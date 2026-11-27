<script>
  // Indikator loading tunggal: empat bola kecil berwarna yang bounce bergiliran.
  let { label = '', size = 18, inline = false, class: className = '' } = $props();
</script>

<span
  class="loading {inline ? 'inline' : 'block'} {className}"
  role="status"
  aria-live="polite"
>
  <span class="dots" style="--dot-size:{size}px" aria-hidden="true">
    <i></i><i></i><i></i><i></i>
  </span>
  {#if label}<span class="label">{label}</span>{/if}
</span>

<style>
  .loading {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    color: var(--muted);
    font: 500 14px/1.4 'DM Mono', monospace;
  }

  .loading.block {
    flex-direction: column;
    gap: 14px;
  }

  .loading.inline {
    gap: 8px;
    color: inherit;
    font: inherit;
  }

  .dots {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 3px;
    flex: none;
    width: var(--dot-size, 18px);
    height: var(--dot-size, 18px);
  }

  .dots i {
    border-radius: 50%;
    animation: load-bounce 1s ease-in-out infinite;
  }

  .dots i:nth-child(1) { background: var(--blue); animation-delay: 0s; }
  .dots i:nth-child(2) { background: var(--red); animation-delay: 0.08s; }
  .dots i:nth-child(3) { background: var(--yellow); animation-delay: 0.16s; }
  .dots i:nth-child(4) { background: var(--green); animation-delay: 0.24s; }

  @keyframes load-bounce {
    0%, 100% { transform: translateY(0); opacity: 0.7; }
    50% { transform: translateY(-5px); opacity: 1; }
  }

  @media (prefers-reduced-motion: reduce) {
    .dots i { animation: none; }
  }
</style>
