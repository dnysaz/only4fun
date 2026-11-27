<script>
  import { api } from './api.js';
  import { sanitize, validateName } from './subdomain.js';
  import { buildTreeRows, normalizeTreePaths } from './fileTree.js';
  import { listZipPaths } from './zipTree.js';
  import { session, signOut } from './session.svelte.js';

  let { preset = '', onCreated } = $props();

  let tab = $state('repo');
  let name = $state('');
  // Nama yang diisi otomatis dari repo picker, supaya bisa diganti saat pindah repo.
  let autoName = $state('');
  let repoUrl = $state('');
  let file = $state(null);
  let fileInput = $state(null);

  let submitting = $state(false);
  let error = $state('');
  let done = $state('');

  let check = $state({ name: '', available: null, error: '', checking: false });
  let tree = $state({ paths: [], loading: false, error: '', truncated: false });
  let treeToken = 0;

  // GitHub repository picker (only for users signed in with GitHub).
  let repoSearch = $state('');
  let repoList = $state([]);
  let repoLoading = $state(false);
  let repoLoadingMore = $state(false);
  let repoError = $state('');
  let repoPage = 1;
  let repoHasMore = $state(false);
  let loadedFor = '';

  // Optional build-time environment variables.
  let envOpen = $state(false);
  let envRows = $state([]);

  // Terms checklist, same pattern as the instant launch page.
  let accepted = $state(false);

  const emptyTree = { paths: [], loading: false, error: '', truncated: false };
  const isValidEnvKey = (key) => /^[A-Za-z_][A-Za-z0-9_]*$/.test(key) && key.length <= 256;

  let nameCheck = $derived(validateName(name));
  let sourceValid = $derived(tab === 'zip' ? !!file : /github\.com/i.test(repoUrl.trim()));
  let treeRows = $derived(buildTreeRows(tree.paths));
  let fileCount = $derived(normalizeTreePaths(tree.paths).length);
  let filteredRepos = $derived(
    repoList.filter((repo) =>
      `${repo.full_name} ${repo.name}`.toLowerCase().includes(repoSearch.trim().toLowerCase()),
    ),
  );
  let envCount = $derived(envRows.filter((row) => row.key.trim()).length);
  let envValid = $derived(
    envRows.every((row) => !row.key.trim() || isValidEnvKey(row.key.trim())),
  );
  let canSubmit = $derived(
    !submitting &&
      accepted &&
      nameCheck.ok &&
      check.available !== false &&
      (tab === 'zip' ? !!file : !!repoUrl.trim() && envValid),
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

  // Load the signed-in user's repositories once when the repo tab is active.
  $effect(() => {
    const user = session.user?.username ?? '';
    if (tab !== 'repo' || !user || loadedFor === user) return;
    loadedFor = user;
    loadRepos();
  });

  async function loadRepos({ append = false } = {}) {
    if (repoLoading || repoLoadingMore || (append && !repoHasMore)) return;
    if (append) repoLoadingMore = true;
    else repoLoading = true;
    repoError = '';

    const nextPage = append ? repoPage + 1 : 1;
    try {
      const data = await api.listGitHubRepos({ page: nextPage, perPage: 100 });
      const incoming = Array.isArray(data?.repos) ? data.repos : [];
      if (append) {
        const seen = new Set(repoList.map((repo) => repo.full_name));
        repoList = [...repoList, ...incoming.filter((repo) => !seen.has(repo.full_name))];
      } else {
        repoList = incoming;
      }
      repoPage = data?.page ?? nextPage;
      repoHasMore = !!data?.has_more && incoming.length > 0;
    } catch (err) {
      // Token GitHub yang sudah dicabut bikin endpoint repo balas 502. Jangan
      // tampilkan error itu di bawah kolom search: langsung minta login ulang.
      const needsLogin =
        err?.status === 401 ||
        (err?.status === 502 && /sign in again/i.test(err?.message ?? ''));
      if (needsLogin) {
        repoList = [];
        repoHasMore = false;
        loadedFor = '';
        repoError = '';
        await signOut();
        window.location.assign(api.loginUrl('/deploy'));
        return;
      }
      repoError = err?.message ?? 'Could not load your repositories';
    } finally {
      repoLoading = false;
      repoLoadingMore = false;
    }
  }

  function repoSelected(repo) {
    const value = repoUrl.trim().toLowerCase().replace(/\.git$/, '');
    const main = `https://github.com/${repo.full_name}`.toLowerCase();
    return value === main || value === (repo.html_url ?? '').toLowerCase();
  }

  function chooseRepo(repo) {
    repoUrl = repo.html_url || `https://github.com/${repo.full_name}`;
    error = '';
    done = '';

    // Ikuti pilihan repo selama nama masih kosong atau masih hasil auto-fill.
    // Kalau user sudah mengetik sendiri, nama manualnya tidak ditimpa.
    if (!name || name === autoName) {
      const suggested = sanitize(repo.name).replace(/^-+|-+$/g, '');
      name = suggested;
      autoName = suggested;
    }
  }

  function addEnv() {
    envRows = [...envRows, { key: '', value: '' }];
  }

  function removeEnv(index) {
    envRows = envRows.filter((_, i) => i !== index);
  }

  function indent(depth) {
    return `${depth * 18 + 12}px`;
  }

  function onName(event) {
    name = sanitize(event.currentTarget.value);
    autoName = '';
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
    if (next === 'zip') {
      // ZIP belum punya nama sampai di-upload atau diketik user.
      name = '';
      autoName = '';
    }
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
      error = 'Paste the GitHub repository URL or pick one below.';
      return;
    }
    if (tab === 'repo' && !envValid) {
      error =
        'Environment names use letters, numbers and underscores, starting with a letter or underscore.';
      return;
    }
    if (!accepted) {
      error = 'Accept the terms and conditions to start the deploy.';
      return;
    }

    const envVars =
      tab === 'repo'
        ? Object.fromEntries(envRows.filter((row) => row.key.trim()).map((row) => [row.key.trim(), row.value]))
        : undefined;

    submitting = true;
    try {
      const data =
        tab === 'zip'
          ? await api.uploadZip(name, file)
          : await api.createProject({ repoUrl: repoUrl.trim(), subdomain: name, envVars });

      done = `${data?.project?.subdomain ?? name}.only4.fun is queued. Build starts in a moment.`;
      repoUrl = '';
      file = null;
      envRows = [];
      envOpen = false;
      accepted = false;
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
            <p class="hint">Paste a URL or pick one of your repositories below. Public and private GitHub repos are supported.</p>

            {#if session.user}
              <div class="repo-picker" aria-label="Your GitHub repositories">
                <div class="picker-head">
                  <span>Your repositories</span>
                  <span class="picker-count">
                    {repoLoading ? 'Loading…' : `${repoList.length} available`}
                  </span>
                </div>

                <div class="picker-search">
                  <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
                    <circle cx="11" cy="11" r="7" />
                    <path d="m20 20-3.5-3.5" />
                  </svg>
                  <input
                    type="search"
                    bind:value={repoSearch}
                    placeholder="Search repositories…"
                    aria-label="Search your repositories"
                    autocomplete="off"
                    spellcheck="false"
                  />
                </div>

                {#if repoError}
                  <p class="picker-error" role="status">
                    {repoError}
                    <button type="button" class="quiet" onclick={() => loadRepos()}>Retry</button>
                  </p>
                {:else if repoLoading}
                  <p class="picker-empty">Loading your repositories…</p>
                {:else if repoList.length === 0}
                  <p class="picker-empty">No repositories found on this GitHub account.</p>
                {:else if filteredRepos.length === 0}
                  <p class="picker-empty">
                    No match in loaded repositories.
                    {#if repoHasMore}
                      <button type="button" class="quiet" onclick={() => loadRepos({ append: true })} disabled={repoLoadingMore}>
                        {repoLoadingMore ? 'Loading…' : 'Load more'}
                      </button>
                    {/if}
                  </p>
                {:else}
                  <ul class="repo-results">
                    {#each filteredRepos as repo (repo.full_name)}
                      <li>
                        <button
                          type="button"
                          class="quiet"
                          class:selected={repoSelected(repo)}
                          onclick={() => chooseRepo(repo)}
                          aria-pressed={repoSelected(repo)}
                        >
                          <svg viewBox="0 0 24 24" width="17" height="17" fill="currentColor" aria-hidden="true">
                            <path d="M12 2a10 10 0 0 0-3.16 19.49c.5.09.68-.22.68-.48v-1.69c-2.78.6-3.37-1.34-3.37-1.34-.45-1.16-1.11-1.47-1.11-1.47-.91-.62.07-.61.07-.61 1 .07 1.53 1.03 1.53 1.03.9 1.53 2.34 1.09 2.91.83.09-.65.35-1.09.64-1.34-2.22-.25-4.55-1.11-4.55-4.94 0-1.09.39-1.98 1.03-2.68-.1-.25-.45-1.27.1-2.64 0 0 .84-.27 2.75 1.02a9.6 9.6 0 0 1 5 0c1.91-1.29 2.75-1.02 2.75-1.02.55 1.37.2 2.39.1 2.64.64.7 1.03 1.59 1.03 2.68 0 3.84-2.34 4.68-4.57 4.93.36.31.68.92.68 1.85v2.74c0 .27.18.58.69.48A10 10 0 0 0 12 2z" />
                          </svg>
                          <span class="repo-name">{repo.name}</span>
                          <span class="repo-full">{repo.full_name}</span>
                          {#if repo.private}<span class="private-tag">Private</span>{/if}
                        </button>
                      </li>
                    {/each}
                  </ul>
                  {#if repoHasMore}
                    <button type="button" class="quiet load-more" onclick={() => loadRepos({ append: true })} disabled={repoLoadingMore}>
                      {repoLoadingMore ? 'Loading more…' : 'Load more repositories'}
                    </button>
                  {/if}
                {/if}
              </div>
            {/if}
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

        {#if tab === 'repo'}
          <div class="env-block">
            <button
              type="button"
              class="quiet env-toggle"
              aria-expanded={envOpen}
              onclick={() => (envOpen = !envOpen)}
            >
              <span class="env-title">Environment variables <span class="env-optional">optional</span></span>
              <span class="env-summary">
                {envCount ? `${envCount} added` : 'Skip for now'}
              </span>
              <svg class="chevron" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="m6 9 6 6 6-6" />
              </svg>
            </button>

            {#if envOpen}
              <div class="env-body">
                <p class="hint">Only needed when your build reads keys such as API_URL. You can add these later in project settings.</p>
                {#each envRows as row, i}
                  <div class="env-row">
                    <input
                      bind:value={row.key}
                      placeholder="API_KEY"
                      aria-label="Variable name"
                      autocomplete="off"
                      spellcheck="false"
                      class:invalid={!!row.key.trim() && !isValidEnvKey(row.key.trim())}
                    />
                    <input
                      bind:value={row.value}
                      placeholder="value"
                      aria-label="Variable value"
                      autocomplete="off"
                      spellcheck="false"
                    />
                    <button type="button" class="quiet env-remove" aria-label="Remove variable" onclick={() => removeEnv(i)}>×</button>
                  </div>
                {/each}
                {#if !envValid}
                  <p class="picker-error">Use letters, numbers and underscores, starting with a letter or underscore.</p>
                {/if}
                <button type="button" class="quiet add-env" onclick={addEnv}>+ Add variable</button>
              </div>
            {/if}
          </div>
        {/if}

        <div class="terms" class:accepted>
          <div class="terms-actions">
            <button
              type="button"
              class="accept"
              aria-pressed={accepted}
              onclick={() => {
                accepted = !accepted;
                error = '';
              }}
            >
              <span class="box" aria-hidden="true">{accepted ? '✓' : ''}</span>
              {accepted ? 'Terms accepted' : 'Accept terms and conditions'}
            </button>
            <a class="terms-link" href="/terms">Terms And Conditions</a>
          </div>
          <p>
            I understand this project will be publicly accessible and build scripts from my repository
            will run on the server.
          </p>
        </div>

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

  /* Keep secondary controls flat; the global blue hover is for primary actions. */
  :global(body .workspace .create button.quiet:hover:not(:disabled)) {
    background: var(--soft) !important;
    border-color: var(--line) !important;
    color: var(--ink) !important;
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

  .repo-picker {
    margin-top: 18px;
    padding-top: 18px;
    border-top: 1px solid var(--line);
  }

  .picker-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    font: 600 16px/1 'Gabarito', sans-serif;
  }

  .picker-count {
    font: 500 13px/1 'DM Mono', monospace;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--muted);
  }

  .picker-search {
    position: relative;
    margin-top: 12px;
  }

  .picker-search svg {
    position: absolute;
    left: 16px;
    top: 50%;
    transform: translateY(-50%);
    color: var(--muted);
  }

  .picker-search input {
    display: block;
    width: 100%;
    border: 1.5px solid var(--line);
    border-radius: 14px;
    background: var(--bg);
    color: var(--ink);
    font: 500 17px/1.3 'Gabarito', sans-serif;
    padding: 13px 16px 13px 44px;
  }

  .picker-search input:focus {
    border-color: var(--blue);
    outline: none;
  }

  .repo-results {
    list-style: none;
    margin: 12px 0 0;
    padding: 0;
    max-height: 190px;
    overflow: auto;
    border: 1px solid var(--line);
    border-radius: 14px;
  }

  .repo-results li + li {
    border-top: 1px solid var(--line);
  }

  .repo-results button {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 9px;
    min-width: 0;
    border: 0 !important;
    border-radius: 0;
    background: transparent;
    padding: 11px 14px;
    text-align: left;
    font: 500 16px/1.3 'Gabarito', sans-serif;
    color: var(--ink);
  }

  .repo-results button svg {
    flex: none;
    color: var(--muted);
  }

  .repo-results button.selected {
    background: var(--soft);
  }

  .repo-results button.selected .repo-name {
    color: var(--blue);
  }

  .repo-name {
    flex: none;
    font-weight: 700;
  }

  .repo-full {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--muted);
    font-size: 15px;
  }

  .private-tag,
  .env-optional {
    flex: none;
    padding: 3px 8px;
    border: 1px solid var(--line);
    border-radius: 999px;
    font: 500 12px/1 'DM Mono', monospace;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--muted);
  }

  .private-tag {
    margin-left: auto;
  }

  .picker-empty,
  .picker-error {
    margin: 14px 0 0;
    font: 500 15px/1.5 'Gabarito', sans-serif;
    color: var(--muted);
  }

  .picker-error {
    color: var(--red);
  }

  .picker-empty button,
  .picker-error button {
    border: 0 !important;
    background: transparent;
    padding: 0 4px;
    color: var(--blue);
    font: 600 15px/1 'Gabarito', sans-serif;
    text-decoration: underline;
  }

  .load-more {
    margin-top: 12px;
    border: 1px solid var(--line) !important;
    background: transparent;
    border-radius: 999px;
    padding: 9px 16px;
    font: 600 15px/1 'Gabarito', sans-serif;
    color: var(--muted);
  }

  .env-block {
    margin-top: 24px;
    padding-top: 22px;
    border-top: 1px solid var(--line);
  }

  .env-toggle {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 10px;
    min-height: 58px;
    border: 0 !important;
    border-radius: 14px;
    background: transparent;
    padding: 14px 16px;
    text-align: left;
    font: 600 17px/1 'Gabarito', sans-serif;
    color: var(--ink);
  }

  .env-title {
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }

  .env-summary {
    margin-left: auto;
    font: 500 15px/1 'Gabarito', sans-serif;
    color: var(--muted);
  }

  .chevron {
    flex: none;
    color: var(--muted);
    transition: transform 0.2s ease;
  }

  .env-toggle[aria-expanded='true'] .chevron {
    transform: rotate(180deg);
  }

  .env-body {
    margin-top: 12px;
  }

  .env-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
    align-items: center;
    gap: 8px;
    margin-top: 10px;
  }

  .env-row input {
    min-width: 0;
    border: 1.5px solid var(--line);
    border-radius: 12px;
    background: var(--bg);
    color: var(--ink);
    font: 500 15px/1 'DM Mono', monospace;
    padding: 12px 13px;
  }

  .env-row input:focus {
    border-color: var(--blue);
    outline: none;
  }

  .env-row input.invalid {
    border-color: var(--red);
  }

  .env-remove {
    border: 0 !important;
    background: transparent;
    padding: 0 4px;
    font: 500 22px/1 'Gabarito', sans-serif;
    color: var(--muted);
  }

  .add-env {
    margin-top: 12px;
    border: 1px solid var(--line) !important;
    background: transparent;
    border-radius: 999px;
    padding: 9px 16px;
    font: 600 15px/1 'Gabarito', sans-serif;
    color: var(--muted);
  }

  .terms {
    margin-top: 24px;
    padding-top: 22px;
    border-top: 1px solid var(--line);
  }

  .terms-actions {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 18px;
  }

  .accept {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    border: 1.5px solid var(--muted);
    border-radius: 999px;
    background: var(--bg);
    color: var(--ink);
    font: 600 16px/1 'Gabarito', sans-serif;
    padding: 12px 20px;
    cursor: pointer;
  }

  .accept:hover:not(:disabled) {
    background: var(--soft) !important;
    border-color: var(--blue) !important;
    color: var(--ink) !important;
  }

  .accept.accepted {
    border-color: var(--blue);
    background: var(--soft);
  }

  .box {
    display: grid;
    place-items: center;
    width: 17px;
    height: 17px;
    border: 1.5px solid currentColor;
    border-radius: 4px;
    font-size: 13px;
    line-height: 1;
  }

  .terms-link {
    color: var(--ink);
    font: 600 16px/1 'Gabarito', sans-serif;
    text-decoration: underline;
    text-decoration-thickness: 1.5px;
    text-underline-offset: 0.2em;
  }

  .terms-link:hover {
    color: var(--blue);
  }

  .terms p {
    max-width: 38em;
    margin: 12px 0 0;
    font-size: 15px;
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
    /* Do not let the file list grow the card; fill the row decided by the form. */
    height: 0;
    min-height: 100%;
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
    flex: 1;
    min-height: 0;
    overflow: auto;
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
      height: auto;
      min-height: 280px;
      margin-top: 26px;
      padding: 22px 0 0;
      border-left: 0;
      border-top: 1px solid var(--line);
    }
  }

  @media (max-width: 700px) {
    .panel {
      padding: 20px 16px;
      border-radius: 18px;
    }

    .bar-top {
      gap: 12px;
      padding-bottom: 16px;
    }

    .note {
      font-size: 12px;
    }

    .tabs {
      width: 100%;
      gap: 6px;
    }

    .tabs button {
      flex: 1;
      min-width: 0;
      padding: 11px 8px;
      font-size: 15px;
    }

    .editor {
      margin-top: 22px;
    }

    .field input,
    .picker-search input {
      min-height: 48px;
      font-size: 16px;
      padding: 13px 14px;
    }

    .picker-search input {
      padding-left: 42px;
    }

    .hint,
    .picker-empty,
    .picker-error,
    .terms p {
      font-size: 14px;
      line-height: 1.55;
    }

    .repo-picker,
    .env-block,
    .name,
    .terms {
      margin-top: 20px;
      padding-top: 20px;
    }

    .repo-results {
      max-height: 220px;
    }

    .repo-results button {
      padding: 12px;
      font-size: 15px;
    }

    .repo-full {
      font-size: 14px;
    }

    .private-tag,
    .env-optional {
      font-size: 11px;
    }

    .env-toggle {
      min-height: 54px;
      padding: 12px;
      font-size: 16px;
    }

    .env-summary {
      font-size: 14px;
    }

    .env-row {
      grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 24px;
      gap: 6px;
    }

    .env-row input {
      min-height: 44px;
      font-size: 13px;
      padding: 11px 9px;
    }

    .env-remove {
      min-height: 44px;
    }

    .terms-actions {
      align-items: stretch;
      gap: 12px;
    }

    .accept {
      width: 100%;
      min-height: 48px;
      justify-content: center;
      padding: 0 16px;
      font-size: 15px;
    }

    .suffix {
      font-size: 15px;
    }

    .status {
      align-items: flex-start;
      font-size: 15px;
      line-height: 1.45;
    }

    .status .dot {
      margin-top: 5px;
    }

    .submit {
      width: 100%;
      min-height: 52px;
      justify-content: center;
      margin-top: 22px;
      font-size: 17px;
    }

    .tree-panel {
      min-height: 0;
      margin-top: 22px;
      padding: 20px 0 0;
    }

    .tree-head,
    .tree-count,
    .tree-root,
    .tree-empty,
    .tree-error,
    .tree-note {
      font-size: 15px;
    }
  }
</style>
