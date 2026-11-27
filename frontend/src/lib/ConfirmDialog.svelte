<script>
  import { tick } from 'svelte';
  import { trapTab } from './focusTrap.js';

  let {
    open = false,
    title = 'Are you sure?',
    message = '',
    confirmLabel = 'Delete',
    cancelLabel = 'Cancel',
    busyLabel = 'Deleting…',
    busy = false,
    tone = 'danger',
    icon = '!',
    showConfirm = true,
    onConfirm = () => {},
    onCancel = () => {},
  } = $props();

  // Read only inside async callbacks/event handlers so the focus effect
  // depends on `open` alone, not on bind:this updates.
  let confirmButton = $state(null);
  let dialog = $state(null);

  $effect(() => {
    if (!open) return;
    const opener = document.activeElement;
    let cancelled = false;

    tick().then(() => {
      if (!cancelled) (confirmButton ?? dialog)?.focus?.();
    });

    return () => {
      cancelled = true;
      if (opener?.isConnected) queueMicrotask(() => opener.focus());
    };
  });

  function keydown(event) {
    if (!open) return;
    if (event.key === 'Escape' && !busy) {
      event.preventDefault();
      onCancel();
      return;
    }
    trapTab(event, dialog);
  }
</script>

<svelte:window onkeydown={keydown} />

{#if open}
  <div class="layer">
    <button
      type="button"
      class="scrim"
      aria-label="Close dialog"
      disabled={busy}
      onclick={onCancel}
    ></button>

    <div
      class="dialog {tone}"
      bind:this={dialog}
      tabindex="-1"
      role="dialog"
      aria-modal="true"
      aria-labelledby="confirm-title"
      aria-describedby="confirm-message"
    >
      <span class="icon" aria-hidden="true">{icon}</span>
      <h2 id="confirm-title">{title}</h2>
      <p id="confirm-message">{message}</p>

      <div class="actions">
        <button type="button" class="cancel" disabled={busy} onclick={onCancel}>{cancelLabel}</button>
        {#if showConfirm}
          <button
            type="button"
            class="confirm"
            bind:this={confirmButton}
            disabled={busy}
            onclick={onConfirm}
          >
            {busy ? busyLabel : confirmLabel}
          </button>
        {/if}
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
    --accent: var(--red);
    position: relative;
    width: min(100%, 460px);
    padding: 34px;
    border: 2px solid var(--line);
    border-radius: 28px;
    background: var(--bg);
  }

  .dialog.success {
    --accent: var(--green);
  }

  .dialog:focus {
    outline: none;
  }

  .icon {
    display: grid;
    place-items: center;
    width: 48px;
    height: 48px;
    margin-bottom: 22px;
    border-radius: 50%;
    background: var(--accent);
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
    white-space: pre-wrap;
    overflow-wrap: anywhere;
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

  /* Override the global blue hover used for primary buttons in the app shell. */
  button:hover:not(:disabled) {
    background: transparent !important;
    border-color: var(--ink) !important;
    color: var(--ink) !important;
  }

  .scrim:hover:not(:disabled) {
    background: color-mix(in srgb, var(--ink) 52%, transparent) !important;
    border-color: transparent !important;
  }

  .confirm {
    border-color: var(--accent);
    background: var(--accent);
    color: #fff;
  }

  .confirm:hover:not(:disabled) {
    background: var(--accent) !important;
    border-color: var(--ink) !important;
    color: #fff !important;
  }

  button:disabled {
    opacity: 0.6;
    cursor: progress;
  }

  .scrim:disabled {
    opacity: 1;
  }
</style>
