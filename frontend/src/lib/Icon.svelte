<script>
  let { name = 'world' } = $props();
</script>

<svg
  viewBox="0 0 24 24"
  width="22"
  height="22"
  fill="none"
  stroke="currentColor"
  stroke-width="1.7"
  stroke-linecap="round"
  stroke-linejoin="round"
  aria-hidden="true"
>
  {#if name === 'world'}
    <circle cx="12" cy="12" r="9" />
    <path d="M3.6 9h16.8" />
    <path d="M3.6 15h16.8" />
    <path d="M11.5 3a17 17 0 0 0 0 18" />
    <path d="M12.5 3a17 17 0 0 1 0 18" />
  {:else if name === 'shield'}
    <path d="M12 3a12 12 0 0 0 8.5 3a12 12 0 0 1-8.5 15a12 12 0 0 1-8.5-15a12 12 0 0 0 8.5-3" />
    <path d="M9 12l2 2l4-4" />
  {:else if name === 'branch'}
    <circle cx="7" cy="6" r="2" />
    <circle cx="7" cy="18" r="2" />
    <circle cx="17" cy="6" r="2" />
    <path d="M7 8v8" />
    <path d="M9 18h6a4 4 0 0 0 4-4V8" />
  {:else if name === 'server'}
    <path d="M3 4h18v7H3z" />
    <path d="M3 13h18v7H3z" />
    <path d="M7 7.5h.01" />
    <path d="M7 16.5h.01" />
  {:else if name === 'bolt'}
    <path d="M13 3v7h5l-7 11v-7H7z" />
  {:else if name === 'infinity'}
    <path
      d="M9.8 9.2a4 4 0 1 0 0 5.6a10 10 0 0 0 2.2-2.8a10 10 0 0 1 2.2-2.8a4 4 0 1 1 0 5.6a10 10 0 0 1-2.2-2.8a10 10 0 0 0-2.2-2.8"
    />
  {/if}
</svg>
