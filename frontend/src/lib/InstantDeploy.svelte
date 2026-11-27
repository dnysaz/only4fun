<script>
  import { api } from './api.js';
  import { Loading } from './loading.ts';
  import { navigate } from './router.svelte.js';
  import Nav from './Nav.svelte';
  import Footer from './Footer.svelte';
  import { listZipPaths } from './zipTree.js';
  import { buildTreeRows, detectFramework, normalizeTreePaths } from './fileTree.js';

  let tab = $state('repo');
  let repoUrl = $state('');
  let file = $state(null);
  let fileInput = $state(null);
  let submitting = $state(false);
  let error = $state('');
  let result = $state(null);
  let copied = $state(false);
  let addressCopied = $state(false);
  let accepted = $state(false);

  let preview = $state({ subdomain: '', url: '', token: '', loading: false });
  let previewToken = 0;
  let tree = $state({ paths: [], signals: [], loading: false, error: '', truncated: false });
  let treeToken = 0;

  const emptyTree = { paths: [], signals: [], loading: false, error: '', truncated: false };

  // Tombol launch ditahan sampai alamat preview selesai dibuat, supaya alamat
  // yang terlihat di form sama persis dengan alamat situs yang di-publish.
  let sourceValid = $derived(tab === 'zip' ? !!file : /github\.com/i.test(repoUrl.trim()));
  let treeRows = $derived(buildTreeRows(tree.paths));
  let fileCount = $derived(normalizeTreePaths(tree.paths).length);
  let frameworkDetected = $derived(detectFramework(tree.paths, tree.signals));
  let launchDisabled = $derived(
    submitting ||
      !sourceValid ||
      tree.loading ||
      !!tree.error ||
      !accepted ||
      frameworkDetected ||
      (sourceValid && preview.loading),
  );

  async function requestPreview() {
    const token = ++previewToken;
    addressCopied = false;
    preview = { subdomain: '', url: '', token: '', loading: true };
    try {
      const data = await api.instantAddress();
      if (token !== previewToken) return;
      preview = {
        subdomain: data?.subdomain ?? '',
        url: data?.url ?? '',
        token: data?.token ?? '',
        loading: false,
      };
    } catch {
      if (token !== previewToken) return;
      preview = { subdomain: '', url: '', token: '', loading: false };
    }
  }

  $effect(() => {
    if (result) return;
    const value = tab === 'repo' ? repoUrl.trim() : (file?.name ?? '');
    const ready = value && (tab === 'zip' || /github\.com/i.test(value));
    if (!ready) {
      previewToken++;
      addressCopied = false;
      preview = { subdomain: '', url: '', token: '', loading: false };
      return;
    }
    const timer = setTimeout(requestPreview, tab === 'zip' ? 0 : 300);
    return () => clearTimeout(timer);
  });

  $effect(() => {
    if (result) return;
    const source = tab === 'repo' ? repoUrl.trim() : file;
    if (!source || (tab === 'repo' && !/github\.com/i.test(source))) {
      treeToken++;
      tree = emptyTree;
      return;
    }

    const token = ++treeToken;
    tree = { paths: [], signals: [], loading: true, error: '', truncated: false };
    const timer = setTimeout(async () => {
      try {
        const data = tab === 'zip' ? await listZipPaths(file) : await api.instantRepoTree(source);
        if (token !== treeToken) return;
        tree = {
          paths: data?.paths ?? [],
          signals: data?.signals ?? [],
          loading: false,
          error: '',
          truncated: !!data?.truncated,
        };
      } catch (err) {
        if (token !== treeToken) return;
        tree = {
          paths: [],
          signals: [],
          loading: false,
          error: err?.message ?? 'Could not read the file list',
          truncated: false,
        };
      }
    }, tab === 'zip' ? 0 : 300);
    return () => clearTimeout(timer);
  });

  // Setiap perubahan sumber harus disetujui lại sebelum launch.
  $effect(() => {
    tab;
    repoUrl;
    file;
    accepted = false;
  });

  function indent(depth) {
    return `${depth * 18 + 12}px`;
  }

  function goHome(event) {
    event.preventDefault();
    navigate('/');
  }

  function onFile(event) {
    file = event.currentTarget.files?.[0] ?? null;
    error = '';
  }

  function setTab(next) {
    tab = next;
    error = '';
  }

  async function submit(event) {
    event.preventDefault();
    error = '';
    copied = false;

    if (!sourceValid || tree.loading || tree.error) return;
    if (!accepted) {
      error = 'Accept the terms and conditions to launch.';
      return;
    }
    if (frameworkDetected) return;

    submitting = true;
    try {
      const data =
        tab === 'zip'
          ? await api.instantZip(file, preview.subdomain, preview.token)
          : await api.instantRepo(repoUrl.trim(), preview.subdomain, preview.token);
      result = {
        subdomain: data?.project?.subdomain ?? '',
        url: data?.url ?? data?.project?.url ?? '',
      };
    } catch (err) {
      error = err?.message ?? 'Could not launch the site';
    } finally {
      submitting = false;
    }
  }

  async function copyUrl() {
    if (!result?.url) return;
    try {
      await navigator.clipboard.writeText(result.url);
      copied = true;
    } catch {
      copied = false;
    }
  }

  async function copyAddress() {
    const value = preview.url || (preview.subdomain ? `https://${preview.subdomain}.only4.fun` : '');
    if (!value) return;
    try {
      await navigator.clipboard.writeText(value);
      addressCopied = true;
      setTimeout(() => {
        addressCopied = false;
      }, 1800);
    } catch {
      addressCopied = false;
    }
  }

  function reset() {
    result = null;
    repoUrl = '';
    file = null;
    error = '';
    copied = false;
    addressCopied = false;
    accepted = false;
    previewToken++;
    treeToken++;
    preview = { subdomain: '', url: '', token: '', loading: false };
    tree = emptyTree;
    if (fileInput) fileInput.value = '';
  }
</script>

<div class="instant workspace">
  <div class="wrap"><Nav /></div>

  {#if !result}
    <section class="wrap hero">
      <p class="eyebrow"><span class="mark" aria-hidden="true"></span>No account needed</p>
      <h1>Launch it <span class="hl">instantly.</span></h1>
      <p class="lead">
        Paste a public GitHub repo or drop a ZIP. We give you a random 8-character address like
        <span class="sample">asj3jkha.only4.fun</span>. Static HTML, CSS, and JS only — JavaScript
        frameworks need an account.
      </p>

      <div class="panel">
        <div class="tabs" role="tablist" aria-label="Source type">
          <button
            type="button"
            class:active={tab === 'repo'}
            role="tab"
            aria-selected={tab === 'repo'}
            onclick={() => setTab('repo')}>GitHub link</button
          >
          <button
            type="button"
            class:active={tab === 'zip'}
            role="tab"
            aria-selected={tab === 'zip'}
            onclick={() => setTab('zip')}>ZIP file</button
          >
        </div>

        <div class="editor">
          <form class="source" onsubmit={submit}>
            {#if tab === 'repo'}
              <label class="field" for="instant-repo">
                <span>Repository URL</span>
                <input
                  id="instant-repo"
                  type="url"
                  placeholder="https://github.com/owner/repo"
                  bind:value={repoUrl}
                  autocomplete="off"
                  spellcheck="false"
                  required
                />
              </label>
              <p class="hint">Public repositories only. The site is served as-is; no build steps run.</p>
            {:else}
              <label class="field" for="instant-zip">
                <span>ZIP archive</span>
                <input
                  id="instant-zip"
                  bind:this={fileInput}
                  type="file"
                  accept=".zip,application/zip"
                  onchange={onFile}
                  required
                />
              </label>
              <p class="hint">Put index.html at the root (or in one folder inside the ZIP). Maximum 100 MB.</p>
            {/if}

            {#if sourceValid}
              <div class="address">
                <span class="address-label">Your address</span>
                <span class="address-line">
                  <span class="address-value" title={preview.url}
                    >{#if preview.loading}
                      <Loading label="generating…" size={12} inline />
                    {:else if preview.subdomain}
                      {preview.subdomain}.only4.fun
                    {:else}
                      ········.only4.fun
                    {/if}</span
                  >
                  {#if preview.subdomain && !preview.loading}
                    <button
                      type="button"
                      class="copy-address"
                      onclick={copyAddress}
                      aria-label="Copy the generated address"
                    >
                      {#if addressCopied}
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                          <path d="m5 13 4 4L19 7" />
                        </svg>
                        Copied
                      {:else}
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                          <rect x="9" y="9" width="11" height="11" rx="2" />
                          <path d="M5 15V6a2 2 0 0 1 2-2h9" />
                        </svg>
                        Copy
                      {/if}
                    </button>
                  {/if}
                </span>
                <span class="address-note">generated automatically</span>
              </div>
            {/if}

            <div class="terms" class:accepted>
              <div class="terms-actions">
                <button
                  type="button"
                  class="accept"
                  aria-pressed={accepted}
                  disabled={!sourceValid || tree.loading}
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
                I understand this site is public, uses a random address, is static-only, and cannot
                be edited or deleted.
              </p>
            </div>

            {#if frameworkDetected}
              <p class="framework-note">
                This looks like a JavaScript framework project.
                <a href={api.loginUrl('/deploy')}>Sign in to deploy frameworks</a>.
              </p>
            {/if}

            {#if error}
              <p class="error" role="alert">{error}</p>
            {/if}

            <button class="submit" type="submit" disabled={launchDisabled}>
              {#if submitting}
                <Loading label="Launching…" size={14} inline />
              {:else}
                Launch Now!
              {/if}
              {#if !submitting && !launchDisabled}
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
          </form>

          <aside class="tree-panel" aria-label="File preview">
            <div class="tree-head">
              <span>File tree</span>
              <span class="tree-count">
                {#if tree.loading}
                  <Loading label="Reading…" size={12} inline />
                {:else if tree.error}
                  Unavailable
                {:else}
                  {fileCount} files
                {/if}
              </span>
            </div>

            {#if !sourceValid}
              <div class="tree-empty">
                Paste a GitHub link or choose a ZIP to see its files here.
              </div>
            {:else if tree.loading}
              <div class="tree-empty"><Loading label="Reading files…" size={22} /></div>
            {:else if tree.error}
              <p class="tree-error" role="status">{tree.error}</p>
              <p class="tree-note">Launch is disabled until the file list can be read.</p>
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
    </section>
  {:else}
    <section class="wrap done">
      <p class="eyebrow"><span class="mark" aria-hidden="true"></span>It's live</p>
      <h1>Your site is <span class="hl">up.</span></h1>

      <div class="live">
        <span class="live-label">Open at</span>
        <a class="live-url" href={result.url} target="_blank" rel="noopener noreferrer">{result.url}</a>
        <div class="live-actions">
          <button type="button" onclick={copyUrl} aria-label="Copy the site address">
            {#if copied}
              <svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M20 6L9 17l-5-5" />
              </svg>
            {:else}
              <svg viewBox="0 0 24 24" width="19" height="19" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <rect x="9" y="9" width="11" height="11" rx="2" />
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
              </svg>
            {/if}
          </button>
          <a class="btn" href={result.url} target="_blank" rel="noopener noreferrer">Open site</a>
        </div>
      </div>

      <p class="notice">
        This address is random and temporary by design. It cannot be edited or deleted, and no
        statistics are available. Want a permanent name you choose, traffic stats, and full control?
        <a href={api.loginUrl('/deploy')}>Sign in to deploy</a>.
      </p>

      <div class="done-actions">
        <button type="button" onclick={reset}>Deploy another</button>
        <a href="/" onclick={goHome}>Back to home</a>
      </div>
    </section>
  {/if}

  <Footer />
</div>

<style>
  .instant {
    min-height: calc(100vh - var(--nav-h));
    background: var(--bg);
  }

  .done-actions a {
    color: var(--muted);
    text-decoration: none;
    font-size: 17px;
  }

  .done-actions a:hover,
  .notice a:hover {
    color: var(--blue);
  }

  .hero,
  .done {
    /* Menjaga padding horizontal dari .wrap supaya sejajar dengan nav/footer.
       Beri ruang napas sebelum footer agar section terakhir tidak mepet. */
    padding-block: 88px clamp(56px, 7vw, 96px);
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

  h1 {
    font-size: clamp(52px, 7.6vw, 104px);
    line-height: 0.95;
    letter-spacing: -0.045em;
    font-weight: 800;
    margin: 0;
  }

  .hl {
    background: linear-gradient(transparent 62%, rgba(251, 188, 4, 0.55) 62%);
    padding: 0 0.04em;
  }

  .lead {
    margin: 28px 0 0;
    max-width: 42em;
    font-size: clamp(19px, 1.8vw, 25px);
    line-height: 1.45;
    color: var(--muted);
  }

  .sample {
    color: var(--ink);
    text-decoration: underline;
    text-decoration-thickness: 2px;
    text-underline-offset: 0.18em;
  }

  .panel {
    margin-top: 48px;
    border: 1px solid var(--line);
    border-radius: 22px;
    padding: 28px;
    width: 100%;
    background: var(--bg);
  }

  .tabs {
    display: flex;
    gap: 8px;
    border-bottom: 1px solid var(--line);
    padding-bottom: 18px;
  }

  .tabs button {
    border: 1px solid transparent;
    background: transparent;
    border-radius: 999px;
    padding: 10px 20px;
    font: 600 17px/1 'Gabarito', sans-serif;
    color: var(--muted);
  }

  :global(body) .instant .tabs button {
    border-color: transparent;
  }

  :global(body) .instant .tabs button.active {
    border-color: var(--line);
    color: var(--ink);
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

  .field span {
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
    color: var(--muted);
  }

  .address {
    display: flex;
    align-items: baseline;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 28px;
    padding-top: 24px;
    border-top: 1px solid var(--line);
  }

  .address-label,
  .address-note {
    font-size: 15px;
    color: var(--muted);
  }

  .address-line {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    min-width: 0;
  }

  .address-value {
    font: 500 24px/1 'DM Mono', monospace;
    letter-spacing: -0.03em;
  }

  .copy-address {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 7px 13px;
    border: 1px solid var(--line);
    border-radius: 999px;
    background: var(--bg);
    color: var(--muted);
    font: 600 13px/1 'Gabarito', sans-serif;
    cursor: pointer;
    transition:
      color 160ms ease,
      border-color 160ms ease,
      background-color 160ms ease;
  }

  .copy-address svg {
    width: 14px;
    height: 14px;
  }

  .copy-address:hover {
    color: #fff;
    border-color: var(--blue);
    background-color: var(--blue);
  }

  .address-note {
    margin-left: auto;
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

  .accept:disabled {
    opacity: 0.5;
    cursor: not-allowed;
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

  .terms p {
    max-width: 38em;
    margin: 12px 0 0;
    font-size: 15px;
    line-height: 1.5;
    color: var(--muted);
  }

  .framework-note {
    margin: 20px 0 0;
    font-size: 16px;
    color: var(--red);
  }

  .framework-note a {
    color: var(--ink);
    text-decoration: underline;
    text-underline-offset: 0.18em;
  }

  .error {
    margin: 22px 0 0;
    color: var(--red);
    font-size: 17px;
  }

  .submit {
    margin-top: 28px;
    display: inline-flex;
    align-items: center;
    gap: 10px;
    border: 1.5px solid var(--muted);
    border-radius: 999px;
    background: var(--blue);
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

  .live {
    margin-top: 48px;
    border-top: 1px solid var(--line);
    border-bottom: 1px solid var(--line);
    padding: 32px 0;
    max-width: 900px;
  }

  .live-label {
    display: block;
    font-size: 15px;
    color: var(--muted);
    margin-bottom: 12px;
  }

  .live-url {
    display: inline-block;
    font: 600 clamp(28px, 4vw, 48px)/1.1 'Gabarito', sans-serif;
    letter-spacing: -0.03em;
    word-break: break-all;
    text-decoration: none;
  }

  .live-url:hover {
    color: var(--blue);
  }

  .live-actions {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-top: 26px;
  }

  .live-actions button {
    display: grid;
    place-items: center;
    width: 48px;
    height: 48px;
    border-radius: 50%;
    border: 1.5px solid var(--muted);
    background: var(--bg);
    cursor: pointer;
  }

  .notice {
    max-width: 46em;
    margin: 32px 0 0;
    font-size: 18px;
    color: var(--muted);
  }

  .notice a {
    color: var(--ink);
    text-decoration: underline;
    text-underline-offset: 0.18em;
  }

  .done-actions {
    display: flex;
    align-items: center;
    gap: 24px;
    margin-top: 36px;
  }

  .done-actions button {
    border: 1.5px solid var(--muted);
    border-radius: 999px;
    background: var(--bg);
    color: var(--ink);
    font: 600 17px/1 'Gabarito', sans-serif;
    padding: 13px 24px;
    cursor: pointer;
  }

  @media (max-width: 980px) {
    .editor {
      grid-template-columns: 1fr;
    }

    .source {
      padding-right: 0;
      /* Beri ruang di bawah tombol sebelum garis pemisah File tree. */
      padding-bottom: 28px;
    }

    .tree-panel {
      height: auto;
      min-height: 280px;
      padding: 22px 0 0;
      border-left: 0;
      border-top: 1px solid var(--line);
    }
  }

  @media (max-width: 760px) {
    .hero {
      padding-block: 48px 56px;
    }

    .done {
      padding-block: 48px 72px;
    }

    .panel {
      padding: 20px;
    }
  }
</style>
