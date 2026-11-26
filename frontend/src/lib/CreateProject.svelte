<script>
  import { api } from './api.js';
  import { sanitize, validateName } from './subdomain.js';
  import { buildTreeRows, normalizeTreePaths } from './fileTree.js';
  import { listZipPaths } from './zipTree.js';

  let { preset = '', onCreated } = $props();

  let tab = $state('repo');
  let name = $state('');
  let repoUrl = $state('');
  let file = $state(null);
  let fileInput = $state(null);

  let submitting = $state(false);
  let error = $state('');
  let done = $state('');

  let check = $state({ name: '', available: null, error: '', checking: false });
  let tree = $state({ paths: [], loading: false, error: '', truncated: false });
  let treeToken = 0;

  const emptyTree = { paths: [], loading: false, error: '', truncated: false };

  let nameCheck = $derived(validateName(name));
  let sourceValid = $derived(tab === 'zip' ? !!file : /github\.com/i.test(repoUrl.trim()));
  let treeRows = $derived(buildTreeRows(tree.paths));
  let fileCount = $derived(normalizeTreePaths(tree.paths).length);
  let canSubmit = $derived(
    !submitting &&
      nameCheck.ok &&
      check.available !== false &&
      (tab === 'zip' ? !!file : !!repoUrl.trim()),
  );

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

  // File tree preview, same helper used by the instant launch page.
  $effect(() => {
    const source = tab === 'repo' ? repoUrl.trim() : file;
    if (!source || (tab === 'repo' && !/github\.com/i.test(source))) {
      treeToken++;
      tree = emptyTree;
      return;
    }

    const token = ++treeToken;
    tree = { paths: [], loading: true, error: '', truncated: false };
    const timer = setTimeout(async () => {
      try {
        const data = tab === 'zip' ? await listZipPaths(file) : await api.instantRepoTree(source);
        if (token !== treeToken) return;
        tree = {
          paths: data?.paths ?? [],
          loading: false,
          error: '',
          truncated: !!data?.truncated,
        };
      } catch (err) {
        if (token !== treeToken) return;
        tree = {
          paths: [],
          loading: false,
          error: err?.message ?? 'Could not read the file list',
          truncated: false,
        };
      }
    }, tab === 'zip' ? 0 : 300);
    return () => clearTimeout(timer);
  });

  function indent(depth) {
    return `${depth * 18 + 12}px`;
  }

  function onName(event) {
    name = sanitize(event.currentTarget.value);
    done = '';
  }

  function onFile(event) {
    file = event.currentTarget.files?.[0] ?? null;
    done = '';
  }

  function setTab(next) {
    tab = next;
    error = '';
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
      if (fileInput) fileInput.value = '';
      onCreated?.(data?.project);
    } catch (err) {
      error = err?.message ?? 'Could not create the project';
    } finally {
      submitting = false;
    }
  }
</script>

<form class="create" onsubmit={submit}>
  <p class="eyebrow"><span class="mark" aria-hidden="true"></span>New deployment</p>
  <h2>Deploy a new project.</h2>
  <p class="lead">Pick a name, add your files, and we'll build and publish the site for you.</p>

  <div class="panel">
    <div class="bar-top">
      <div class="tabs" role="tablist" aria-label="Source type">
        <button
          type="button"
          role="tab"
          aria-selected={tab === 'repo'}
          class:active={tab === 'repo'}
          onclick={() => setTab('repo')}>GitHub link</button
        >
        <button
          type="button"
          role="tab"
          aria-selected={tab === 'zip'}
          class:active={tab === 'zip'}
          onclick={() => setTab('zip')}>ZIP file</button
        >
      </div>
      <span class="note">free · no limits</span>
    </div>

    <div class="editor">
      <div class="source">
        <label class="field" for="deploy-repo">
          <span>{tab === 'repo' ? 'Repository URL' : 'ZIP archive'}</span>
          {#if tab === 'repo'}
            <input
              id="deploy-repo"
              type="url"
              placeholder="https://github.com/you/repo"
              bind:value={repoUrl}
              autocomplete="off"
              spellcheck="false"
            />
            <p class="hint">Public GitHub repositories. npm, pnpm and yarn builds are supported.</p>
          {:else}
            <input
              id="deploy-repo"
              bind:this={fileInput}
              type="file"
              accept=".zip,application/zip"
              onchange={onFile}
            />
            <p class="hint">Put index.html at the root (or in one folder inside the ZIP). Maximum 100 MB.</p>
          {/if}
        </label>

        <label class="field name" for="site-name">
          <span>SITE NAME</span>
          <span class="input-line">
            <input
              id="site-name"
              value={name}
              oninput={onName}
              placeholder="my-fun-site"
              maxlength={63}
              autocomplete="off"
              spellcheck="false"
            />
            <span class="suffix">.only4.fun</span>
          </span>
        </label>

        <p
          class="status"
          class:ok={check.available === true}
          class:bad={check.available === false || (!nameCheck.ok && !!name)}
        >
          <span class="dot" aria-hidden="true"></span>
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

        {#if error}
          <p class="error" role="alert">{error}</p>
        {/if}
        {#if done}
          <p class="done" role="status">{done}</p>
        {/if}

        <button class="submit" type="submit" disabled={!canSubmit}>
          {submitting ? 'Starting…' : 'Start deploy'}
          {#if !submitting}
            <svg
              viewBox="0 0 24 24"
              width="18"
              height="18"
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
          {/if}
        </button>
      </div>

      <aside class="tree-panel" aria-label="File preview">
        <div class="tree-head">
          <span>File tree</span>
          <span class="tree-count">
            {tree.loading ? 'Reading…' : tree.error ? 'Unavailable' : `${fileCount} files`}
          </span>
        </div>

        {#if !sourceValid}
          <div class="tree-empty">Paste a GitHub link or choose a ZIP to see its files here.</div>
        {:else if tree.loading}
          <div class="tree-empty">Reading files…</div>
        {:else if tree.error}
          <p class="tree-error" role="status">{tree.error}</p>
          <p class="tree-note">The preview is unavailable, but you can still deploy.</p>
        {:else}
          <div class="tree-root">
            <svg viewBox="0 0 24 24" width="18" height="18" fill="currentColor" aria-hidden="true">
              <path d="M3 6.5A2.5 2.5 0 0 1 5.5 4H9l2 2h7.5A2.5 2.5 0 0 1 21 8.5v9a2.5 2.5 0 0 1-2.5 2.5h-13A2.5 2.5 0 0 1 3 17.5z" />
            </svg>
            site/
          </div>
          <ul class="tree-list">
            {#each treeRows as row (row.path)}
              <li class={row.kind} style:padding-left={indent(row.depth)}>
                {#if row.kind === 'folder'}
                  <svg class="folder-icon" viewBox="0 0 24 24" width="18" height="18" fill="currentColor" aria-hidden="true">
                    <path d="M3 6.5A2.5 2.5 0 0 1 5.5 4H9l2 2h7.5A2.5 2.5 0 0 1 21 8.5v9a2.5 2.5 0 0 1-2.5 2.5h-13A2.5 2.5 0 0 1 3 17.5z" />
                  </svg>
                {:else if row.kind === 'file'}
                  <svg class="file-icon" viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="M6 3h8l4 4v14H6z" />
                    <path d="M14 3v5h4" />
                    <path d="M9 13h6M9 17h4" />
                  </svg>
                {:else}
                  <span class="more-dot" aria-hidden="true">···</span>
                {/if}
                <span class="tree-name">{row.name}</span>
                {#if row.kind === 'file' && row.name === 'index.html'}
                  <span class="entry-tag">entry</span>
                {/if}
              </li>
            {/each}
          </ul>
          {#if tree.truncated}
            <p class="tree-note">Showing a limited preview of the files.</p>
          {/if}
        {/if}
      </aside>
    </div>
  </div>
</form>

<style>
  .create {
    padding-bottom: 36px;
  }

  .eyebrow {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    margin: 0 0 26px;
    padding: 8px 16px;
    border: 1px solid var(--line);
    border-radius: 999px;
    font: 500 14px/1 'DM Mono', monospace;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--muted);
  }

  .mark {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    background: var(--yellow);
  }

  h2 {
    margin: 0;
    font-size: clamp(38px, 5vw, 64px);
    line-height: 1;
    letter-spacing: -0.045em;
    font-weight: 800;
  }

  .lead {
    margin: 20px 0 0;
    max-width: 46em;
    font-size: clamp(17px, 1.5vw, 20px);
    line-height: 1.5;
    color: var(--muted);
  }

  .panel {
    margin-top: 36px;
    border: 1px solid var(--line);
    border-radius: 22px;
    padding: 28px;
    background: var(--bg);
  }

  .bar-top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    flex-wrap: wrap;
    padding-bottom: 18px;
    border-bottom: 1px solid var(--line);
  }

  .tabs {
    display: flex;
    gap: 8px;
  }

  .tabs button {
    border: 1px solid transparent !important;
    background: transparent;
    border-radius: 999px;
    padding: 10px 20px;
    font: 600 17px/1 'Gabarito', sans-serif;
    color: var(--muted);
  }

  .tabs button.active {
    border-color: var(--line) !important;
    color: var(--ink);
  }

  .note {
    color: var(--muted);
    font: 500 13px/1 'DM Mono', monospace;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }

  .editor {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    margin-top: 26px;
  }

  .source {
    min-width: 0;
    padding-right: 32px;
  }

  .field {
    display: block;
    font-size: 16px;
    font-weight: 600;
  }

  .field > span:first-child {
    display: block;
    margin-bottom: 10px;
    color: var(--muted);
  }

  .field input {
    display: block;
    width: 100%;
    border: 1.5px solid var(--line);
    border-radius: 14px;
    background: var(--bg);
    color: var(--ink);
    font: 500 19px/1.3 'Gabarito', sans-serif;
    padding: 15px 18px;
  }

  .field input:focus {
    border-color: var(--blue);
    outline: none;
  }

  .field input::file-selector-button {
    border: 1.5px solid var(--muted);
    border-radius: 999px;
    background: var(--soft);
    color: var(--ink);
    font: 600 15px/1 'Gabarito', sans-serif;
    padding: 9px 16px;
    margin-right: 14px;
    cursor: pointer;
  }

  .hint {
    margin: 12px 0 0;
    font-size: 15px;
    font-weight: 400;
    line-height: 1.5;
    color: var(--muted);
  }

  .name {
    margin-top: 28px;
    padding-top: 24px;
    border-top: 1px solid var(--line);
  }

  .input-line {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .input-line input {
    flex: 1;
    min-width: 0;
    font-family: 'DM Mono', monospace;
    letter-spacing: -0.02em;
  }

  .suffix {
    flex: none;
    color: var(--muted);
    font: 500 17px/1 'DM Mono', monospace;
    white-space: nowrap;
  }

  .status {
    display: flex;
    align-items: center;
    gap: 9px;
    margin: 16px 0 0;
    font-size: 16px;
    font-weight: 400;
    color: var(--muted);
  }

  .status .dot {
    flex: none;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--line);
  }

  .status.ok {
    color: var(--green);
  }

  .status.ok .dot {
    background: var(--green);
  }

  .status.bad {
    color: var(--red);
  }

  .status.bad .dot {
    background: var(--red);
  }

  .error,
  .done {
    margin: 16px 0 0;
    font-size: 16px;
    font-weight: 500;
  }

  .error {
    color: var(--red);
  }

  .done {
    color: var(--green);
  }

  .submit {
    margin-top: 28px;
    display: inline-flex;
    align-items: center;
    gap: 10px;
    border-radius: 999px;
    background: var(--blue);
    border: 1.5px solid var(--blue);
    color: #fff;
    font: 600 19px/1 'Gabarito', sans-serif;
    padding: 16px 30px;
    cursor: pointer;
  }

  .submit:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  .tree-panel {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 360px;
    padding-left: 32px;
    border-left: 1px solid var(--line);
  }

  .tree-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding-bottom: 14px;
    border-bottom: 1px solid var(--line);
    font: 600 17px/1 'Gabarito', sans-serif;
  }

  .tree-count {
    font: 500 17px/1 'Gabarito', sans-serif;
    color: var(--muted);
  }

  .tree-empty {
    display: grid;
    place-items: center;
    flex: 1;
    padding: 24px;
    text-align: center;
    font: 500 17px/1.5 'Gabarito', sans-serif;
    color: var(--muted);
  }

  .tree-error {
    margin: 20px 0 0;
    color: var(--red);
    font: 500 17px/1.5 'Gabarito', sans-serif;
  }

  .tree-root {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 16px 0 6px;
    font: 600 17px/1.4 'Gabarito', sans-serif;
  }

  .tree-root svg {
    flex: none;
    color: var(--yellow);
  }

  .tree-list {
    list-style: none;
    margin: 0;
    padding: 0;
    overflow: auto;
    max-height: 330px;
    font: 500 17px/1.7 'Gabarito', sans-serif;
  }

  .tree-list li {
    display: flex;
    align-items: center;
    gap: 9px;
    min-width: max-content;
    white-space: nowrap;
  }

  .tree-list .folder {
    color: var(--ink);
    font-weight: 700;
  }

  .tree-list .file,
  .tree-list .more {
    color: var(--muted);
  }

  .folder-icon {
    flex: none;
    color: var(--yellow);
  }

  .file-icon {
    flex: none;
    color: var(--muted);
  }

  .more-dot {
    width: 18px;
    color: var(--muted);
  }

  .tree-name {
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .entry-tag {
    margin-left: 6px;
    padding: 3px 10px;
    border: 1px solid var(--line);
    border-radius: 999px;
    font: 600 15px/1 'Gabarito', sans-serif;
    color: var(--blue);
  }

  .tree-note {
    margin: auto 0 0;
    padding-top: 12px;
    font: 500 16px/1.5 'Gabarito', sans-serif;
    color: var(--muted);
  }

  @media (max-width: 980px) {
    .editor {
      grid-template-columns: 1fr;
    }

    .source {
      padding-right: 0;
    }

    .tree-panel {
      min-height: 280px;
      margin-top: 26px;
      padding: 22px 0 0;
      border-left: 0;
      border-top: 1px solid var(--line);
    }
  }

  @media (max-width: 700px) {
    .panel {
      padding: 20px;
    }

    .submit {
      width: 100%;
      justify-content: center;
    }
  }
</style>
