<script>
  import { api } from './api.js';
  import { sanitize, validateName } from './subdomain.js';

  let { preset = '', onCreated } = $props();

  let tab = $state('repo');
  let name = $state('');
  let repoUrl = $state('');
  let file = $state(null);

  let submitting = $state(false);
  let error = $state('');
  let done = $state('');

  let check = $state({ name: '', available: null, error: '', checking: false });

  let nameCheck = $derived(validateName(name));

  // Isi nama dari "?claim=" di URL hanya sekali (bukan tiap render).
  let lastPreset = '';
  $effect(() => {
    if (preset && preset !== lastPreset) {
      lastPreset = preset;
      name = sanitize(preset);
    }
  });

  $effect(() => {
    const value = nameCheck.ok ? name : '';
    if (!value) {
      check = { name: '', available: null, error: '', checking: false };
      return;
    }

    let cancelled = false;
    const timer = setTimeout(async () => {
      check = { name: value, available: null, error: '', checking: true };
      try {
        const data = await api.checkSubdomain(value);
        if (cancelled) return;
        check = {
          name: value,
          available: !!data?.available,
          error: data?.available ? '' : (data?.error ?? 'that name is not available'),
          checking: false,
        };
      } catch (err) {
        if (cancelled) return;
        check = {
          name: value,
          available: null,
          error: err?.message ?? 'Could not check the name',
          checking: false,
        };
      }
    }, 400);

    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  });

  function onName(event) {
    name = sanitize(event.currentTarget.value);
    done = '';
  }

  function onFile(event) {
    file = event.currentTarget.files?.[0] ?? null;
    done = '';
  }

  async function submit(event) {
    event.preventDefault();
    error = '';
    done = '';

    if (!nameCheck.ok) {
      error =
        nameCheck.reason === 'blocked'
          ? `"${name}" is not allowed. Pick another name.`
          : nameCheck.reason === 'reserved'
            ? `"${name}" is reserved. Pick another name.`
            : nameCheck.reason === 'short'
              ? 'Use at least 3 characters.'
              : nameCheck.reason === 'hyphen'
              ? "Names can't start or end with a hyphen."
              : 'Pick a name for your site.';
      return;
    }
    if (check.available === false) {
      error = check.error || 'That name is taken.';
      return;
    }
    if (tab === 'zip' && !file) {
      error = 'Choose a .zip file to upload.';
      return;
    }
    if (tab === 'repo' && !repoUrl.trim()) {
      error = 'Paste the GitHub repository URL.';
      return;
    }

    submitting = true;
    try {
      const data =
        tab === 'zip'
          ? await api.uploadZip(name, file)
          : await api.createProject({ repoUrl: repoUrl.trim(), subdomain: name });

      done = `${data?.project?.subdomain ?? name}.only4.fun is queued. Build starts in a moment.`;
      repoUrl = '';
      file = null;
      onCreated?.(data?.project);
    } catch (err) {
      error = err?.message ?? 'Could not create the project';
    } finally {
      submitting = false;
    }
  }
</script>

<form class="create" onsubmit={submit}>
  <p class="eyebrow"><span class="pulse" aria-hidden="true"></span>New deployment</p>
  <h2>Deploy a new project.</h2>

  <div class="bar">
    <div class="bar-top">
      <div class="tabs" role="tablist">
        <button
          type="button"
          role="tab"
          aria-selected={tab === 'repo'}
          class:active={tab === 'repo'}
          onclick={() => (tab = 'repo')}
        >
          GitHub repo
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={tab === 'zip'}
          class:active={tab === 'zip'}
          onclick={() => (tab = 'zip')}
        >
          Upload ZIP
        </button>
      </div>
      <span class="note">free · no limits</span>
    </div>

    <div class="fields">
      <label class="field">
        <span>SITE NAME</span>
        <div class="input">
          <input
            value={name}
            oninput={onName}
            placeholder="my-fun-site"
            maxlength={63}
            autocomplete="off"
            spellcheck="false"
            aria-label="Subdomain"
          />
          <span class="suffix">.only4.fun</span>
        </div>
      </label>

      {#if tab === 'repo'}
        <label class="field">
          <span>GITHUB REPOSITORY</span>
          <input
            class="plain"
            bind:value={repoUrl}
            placeholder="https://github.com/you/repo"
            autocomplete="off"
            spellcheck="false"
            aria-label="GitHub repository URL"
          />
        </label>
      {:else}
        <label class="field">
          <span>STATIC SITE ZIP</span>
          <input
            class="plain file"
            type="file"
            accept=".zip,application/zip"
            onchange={onFile}
            aria-label="ZIP file"
          />
        </label>
      {/if}
    </div>

    <div class="bar-bottom">
      <p class="status" class:ok={check.available === true} class:bad={check.available === false}>
        {#if !name}
          Pick a name — letters, numbers and hyphens.
        {:else if !nameCheck.ok}
          {nameCheck.reason === 'blocked'
            ? 'That name is not allowed.'
            : nameCheck.reason === 'reserved'
              ? 'That name is reserved.'
              : nameCheck.reason === 'short'
                ? 'At least 3 characters.'
                : nameCheck.reason === 'hyphen'
                  ? "Can't start or end with a hyphen."
                  : ''}
        {:else if check.checking}
          Checking availability…
        {:else if check.available === true}
          {name}.only4.fun is available.
        {:else if check.available === false}
          {check.error}
        {:else if check.error}
          {check.error}
        {:else}
          {name}.only4.fun looks good.
        {/if}
      </p>

      <button class="btn deploy" type="submit" disabled={submitting}>
        {submitting ? 'Starting…' : 'Start deploy'}
        <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M7 17L17 7" /><path d="M8 7h9v9" />
        </svg>
      </button>
    </div>
  </div>

  {#if error}
    <p class="status bad below">{error}</p>
  {/if}
  {#if done}
    <p class="status ok below">{done}</p>
  {/if}
</form>

<style>
  .create {
    padding-bottom: 36px;
  }

  .eyebrow {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    margin: 0 0 20px;
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

  h2 {
    margin: 0 0 26px;
    font-size: clamp(34px, 4.6vw, 56px);
    line-height: 1;
    letter-spacing: -0.04em;
    font-weight: 800;
  }

  .bar {
    border: 2px solid var(--line);
    border-radius: 30px;
    padding: 24px 26px;
    background: var(--bg);
    transition: border-color 0.15s;
  }

  .bar:focus-within {
    border-color: var(--blue);
  }

  .bar-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    flex-wrap: wrap;
    margin-bottom: 20px;
  }

  .tabs {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }

  .tabs button {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    padding: 9px 18px 9px 14px;
    border: 2px solid var(--line);
    border-radius: 999px;
    background: transparent;
    color: var(--muted);
    font: 600 16px/1 'Gabarito', sans-serif;
    cursor: pointer;
  }

  .tabs button::before {
    content: '';
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background: var(--line);
  }

  .tabs button.active {
    border-color: var(--blue);
    color: var(--ink);
  }

  .tabs button.active::before {
    background: var(--blue);
  }

  .note {
    color: var(--muted);
    font: 500 13px/1 'DM Mono', monospace;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }

  .fields {
    display: grid;
    grid-template-columns: minmax(220px, 0.9fr) minmax(0, 1.4fr);
    gap: 14px;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 0;
  }

  .field > span {
    color: var(--muted);
    font: 500 12px/1 'DM Mono', monospace;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }

  .input {
    display: flex;
    align-items: center;
    gap: 2px;
    min-height: 60px;
    padding: 0 18px;
    border: 2px solid var(--line);
    border-radius: 16px;
    background: var(--bg);
    transition: border-color 0.15s;
  }

  .input:focus-within {
    border-color: var(--blue);
  }

  .input input {
    flex: 1;
    min-width: 0;
    border: 0;
    outline: 0;
    background: transparent;
    color: var(--ink);
    font: 500 20px/1 'DM Mono', monospace;
    letter-spacing: -0.03em;
  }

  .input input::placeholder {
    color: #bdc1c6;
  }

  .suffix {
    color: var(--muted);
    font: 500 18px/1 'DM Mono', monospace;
    letter-spacing: -0.03em;
    white-space: nowrap;
  }

  .plain {
    width: 100%;
    min-height: 60px;
    padding: 0 18px;
    border: 2px solid var(--line);
    border-radius: 16px;
    outline: 0;
    background: var(--bg);
    color: var(--ink);
    font: 500 17px/1 'Gabarito', sans-serif;
    transition: border-color 0.15s;
  }

  .plain:focus {
    border-color: var(--blue);
  }

  .plain.file {
    font-size: 14px;
    cursor: pointer;
  }

  .bar-bottom {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    flex-wrap: wrap;
    margin-top: 20px;
  }

  .status {
    margin: 0;
    flex: 1 1 260px;
    color: var(--muted);
    font-size: 17px;
  }

  .status.ok { color: var(--green); }
  .status.bad { color: var(--red); }

  .status.below {
    margin-top: 16px;
  }

  .deploy {
    display: inline-flex;
    align-items: center;
    gap: 9px;
    margin: 0;
    padding: 15px 26px;
    font-size: 18px;
  }

  .deploy:disabled {
    cursor: progress;
    opacity: 0.65;
  }

  @media (max-width: 900px) {
    .bar {
      padding: 20px;
    }

    .fields {
      grid-template-columns: 1fr;
    }

    .deploy {
      width: 100%;
      justify-content: center;
    }
  }
</style>
