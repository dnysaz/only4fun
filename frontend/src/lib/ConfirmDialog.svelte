<script>
  let {
    open = false,
    title = 'Are you sure?',
    message = '',
    confirmLabel = 'Delete',
    busyLabel = 'Deleting…',
    busy = false,
    onConfirm = () => {},
    onCancel = () => {},
  } = $props();

  let confirmButton = $state(null);

  $effect(() => {
    if (open) confirmButton?.focus();
  });

  function keydown(event) {
    if (!open || busy || event.key !== 'Escape') return;
    event.preventDefault();
    onCancel();
  }
</script>

<svelte:window onkeydown={keydown} />

{#if open}
  <div class="layer">
    <button
      type="button"
      class="scrim"
      aria-label="Close confirmation"
      disabled={busy}
      onclick={onCancel}
    ></button>

    <div
      class="dialog"
      role="dialog"
      aria-modal="true"
      aria-labelledby="confirm-title"
      aria-describedby="confirm-message"
    >
      <span class="icon" aria-hidden="true">!</span>
      <h2 id="confirm-title">{title}</h2>
      <p id="confirm-message">{message}</p>

      <div class="actions">
        <button type="button" class="cancel" disabled={busy} onclick={onCancel}>Cancel</button>
        <button
          type="button"
          class="confirm"
          bind:this={confirmButton}
          disabled={busy}
          onclick={onConfirm}
        >
          {busy ? busyLabel : confirmLabel}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .layer {
    position: fixed;
    inset: 0;
    z-index: 100;
    display: grid;
    place-items: center;
    padding: 20px;
  }

  .scrim {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    border: 0;
    border-radius: 0;
    background: color-mix(in srgb, var(--ink) 52%, transparent);
    cursor: default;
  }

  .dialog {
    position: relative;
    width: min(100%, 460px);
    padding: 34px;
    border: 2px solid var(--line);
    border-radius: 28px;
    background: var(--bg);
  }

  .icon {
    display: grid;
    place-items: center;
    width: 48px;
    height: 48px;
    margin-bottom: 22px;
    border-radius: 50%;
    background: var(--red);
    color: #fff;
    font-size: 28px;
    font-weight: 800;
    line-height: 1;
  }

  h2 {
    margin: 0 0 12px;
    font-size: clamp(28px, 4vw, 38px);
    line-height: 1.05;
    letter-spacing: -0.04em;
    font-weight: 800;
  }

  p {
    margin: 0 0 28px;
    color: var(--muted);
    font-size: 16px;
    line-height: 1.5;
  }

  .actions {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }

  button {
    padding: 13px 22px;
    border: 2px solid var(--line);
    border-radius: 999px;
    background: transparent;
    color: var(--ink);
    font: 600 16px/1 'Gabarito', sans-serif;
    cursor: pointer;
  }

  button:hover:not(:disabled) {
    border-color: var(--ink);
  }

  .confirm {
    border-color: var(--red);
    background: var(--red);
    color: #fff;
  }

  .confirm:hover:not(:disabled) {
    border-color: var(--ink);
  }

  button:disabled {
    opacity: 0.6;
    cursor: progress;
  }

  .scrim:disabled {
    opacity: 1;
  }
</style>
