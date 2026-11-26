<script>
  import { onDestroy } from 'svelte';
  import Workspace from './Workspace.svelte';
  import ConfirmDialog from './ConfirmDialog.svelte';
  import { api } from './api.js';
  import { session } from './session.svelte.js';
  import { navigate, getRoute } from './router.svelte.js';
  import { reload } from './workspace.svelte.js';

  const KEY = /^[A-Za-z_][A-Za-z0-9_]*$/;
  const CNAME_TARGET = import.meta.env.VITE_CUSTOM_DOMAIN_CNAME || 'customers.only4.fun';
  const STATUS = {
    deployed: { label: 'Live', tone: 'green' },
    building: { label: 'Building', tone: 'yellow' },
    pending: { label: 'Queued', tone: 'yellow' },
    failed: { label: 'Failed', tone: 'red' },
  };
  const DEPLOY_TONE = {
    success: 'green',
    failed: 'red',
    building: 'yellow',
    pending: 'yellow',
  };
  const DOMAIN_STATE_LABEL = {
    active: 'Active',
    ssl_pending: 'Issuing certificate',
    dns_pending: 'Waiting for DNS',
    pending: 'Verifying',
    none: 'No domain',
    ssl_unavailable: 'SSL unavailable',
    unknown: 'Status unknown',
  };
  const DOMAIN_STATE_TONE = {
    active: 'green',
    ssl_pending: 'yellow',
    dns_pending: 'yellow',
    pending: 'yellow',
    ssl_unavailable: 'red',
    unknown: 'red',
  };
  const CHART_H = 180;
  const CHART_PAD = 14;
  let chartWidth = $state(600);
  let chartW = $derived(Math.max(1, chartWidth));

  function smoothPath(points) {
    if (!points.length) return '';
    let d = `M ${points[0].x} ${points[0].y}`;
    for (let i = 0; i < points.length - 1; i += 1) {
      const p0 = points[i - 1] ?? points[i];
      const p1 = points[i];
      const p2 = points[i + 1];
      const p3 = points[i + 2] ?? p2;
      const c1x = p1.x + (p2.x - p0.x) / 6;
      const c1y = p1.y + (p2.y - p0.y) / 6;
      const c2x = p2.x - (p3.x - p1.x) / 6;
      const c2y = p2.y - (p3.y - p1.y) / 6;
      d += ` C ${c1x} ${c1y}, ${c2x} ${c2y}, ${p2.x} ${p2.y}`;
    }
    return d;
  }

  let id = $derived(getRoute().split('/')[2] ?? '');
  let project = $state(null);
  let deployments = $state([]);
  let loading = $state(true);
  let error = $state('');
  let notice = $state('');
  let domainWarning = $state('');
  let domainState = $state('none');
  let domainStatusMessage = $state('');
  let domainStatusTimer = null;
  let checkingDomain = $state(false);
  let busy = $state(false);

  let envRows = $state([]);
  let domain = $state('');
  let domainLabel = $derived.by(() => {
    const parts = domain.trim().toLowerCase().replace(/\.$/, '').split('.').filter(Boolean);
    return parts.length > 2 ? parts[0] : 'www';
  });
  let savingEnv = $state(false);
  let savingDomain = $state(false);
  let showDelete = $state(false);
  let deleting = $state(false);
  let copiedLog = $state('');

  // Live build log: poll the deployment record while a build is running.
  let livePolling = $state(false);
  let liveLogId = $state('');
  let pollTimer = null;

  let loadedFor = '';

  let meta = $derived(STATUS[project?.status] ?? { label: project?.status ?? '—', tone: 'blue' });
  let url = $derived(project ? (project.deploy_url || `https://${project.subdomain}.only4.fun`) : '');
  let repo = $derived(
    project?.repo_url?.startsWith('zip://')
      ? 'Uploaded ZIP'
      : (project?.repo_url ?? '').replace(/^https?:\/\/(www\.)?github\.com\//, ''),
  );

  let visitDays = $derived.by(() => {
    const counts = new Map((project?.visits ?? []).map((v) => [v.day, Number(v.count) || 0]));
    const days = [];
    const today = new Date();
    for (let i = 13; i >= 0; i -= 1) {
      const d = new Date(today);
      d.setDate(today.getDate() - i);
      const key = d.toISOString().slice(0, 10);
      days.push({ day: key, count: counts.get(key) ?? 0 });
    }
    return days;
  });
  let maxVisits = $derived(Math.max(1, ...visitDays.map((v) => v.count)));
  let points = $derived(
    visitDays.map((v, i) => ({
      x: (i / (visitDays.length - 1)) * chartW,
      y:
        CHART_H -
        CHART_PAD -
        (v.count / maxVisits) * (CHART_H - CHART_PAD * 2),
      count: v.count,
      day: v.day,
    })),
  );
  let linePath = $derived(smoothPath(points));
  let areaPath = $derived(`${linePath} L ${chartW} ${CHART_H} L 0 ${CHART_H} Z`);
  let recentViews = $derived(visitDays.reduce((sum, v) => sum + v.count, 0));
  let totalViews = $derived(project?.views ?? 0);

  function fmt(value, withTime = false) {
    if (!value) return '—';
    return new Date(value).toLocaleString(undefined, {
      dateStyle: 'medium',
      timeStyle: withTime ? 'short' : undefined,
    });
  }

  function rowsFromKeys(keys = []) {
    return keys.map((key) => ({ key, value: '', remove: false, existing: true }));
  }

  function isRunning(dep) {
    return dep?.status === 'building' || dep?.status === 'pending';
  }

  function withOpenState(incoming) {
    const previous = new Map(deployments.map((dep) => [dep.id, !!dep.open]));
    return incoming.map((dep) => ({
      ...dep,
      open: previous.get(dep.id) || dep.id === liveLogId || isRunning(dep),
    }));
  }

  async function refreshLive() {
    if (!id) return;
    try {
      const [detail, history] = await Promise.all([api.getProject(id), api.deployments(id)]);
      project = detail?.project ?? project;
      const incoming = history?.deployments ?? [];

      // A redeploy creates its deployment record asynchronously in the worker.
      if (livePolling && !liveLogId) {
        const known = new Set(deployments.map((dep) => dep.id));
        const fresh = incoming.find((dep) => !known.has(dep.id));
        if (fresh) liveLogId = fresh.id;
      }

      deployments = withOpenState(incoming);
      const stillRunning =
        incoming.some(isRunning) ||
        project?.status === 'pending' ||
        project?.status === 'building';

      if (stillRunning) {
        notice = 'Build in progress — the log is updating live.';
      } else if (livePolling) {
        const newest = incoming[0];
        if (newest?.status === 'success') notice = 'Deploy finished successfully.';
        else if (newest?.status === 'failed') {
          notice = 'Deploy failed. Open the build log for details.';
        }
        livePolling = false;
      }
    } catch (err) {
      error = err?.message ?? 'Could not refresh the build log';
      livePolling = false;
    }
  }

  async function load() {
    if (!id) return;
    loading = true;
    error = '';
    try {
      const [detail, history] = await Promise.all([api.getProject(id), api.deployments(id)]);
      project = detail?.project ?? null;
      const incoming = history?.deployments ?? [];
      deployments = withOpenState(incoming);
      if (incoming.some(isRunning) || ['pending', 'building'].includes(project?.status)) {
        livePolling = true;
      }
      envRows = rowsFromKeys(project?.env_keys ?? []);
      domain = project?.custom_domain ?? '';
      loadDomainStatus();
    } catch (err) {
      error = err?.message ?? 'Could not load the project';
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    if (!session.user || !id || loadedFor === id) return;
    loadedFor = id;
    load();
  });

  $effect(() => {
    const running =
      livePolling ||
      deployments.some(isRunning) ||
      project?.status === 'pending' ||
      project?.status === 'building';
    if (running && !pollTimer) {
      pollTimer = window.setInterval(refreshLive, 900);
    } else if (!running && pollTimer) {
      window.clearInterval(pollTimer);
      pollTimer = null;
    }
  });

  onDestroy(() => {
    if (pollTimer) window.clearInterval(pollTimer);
    if (domainStatusTimer) window.clearInterval(domainStatusTimer);
  });

  async function redeploy() {
    busy = true;
    notice = '';
    error = '';
    liveLogId = '';
    livePolling = true;
    try {
      await api.redeploy(id);
      notice = 'Build starting — the log is updating live.';
      await refreshLive();
    } catch (err) {
      error = err?.message ?? 'Could not start a deploy';
      livePolling = false;
    } finally {
      busy = false;
    }
  }

  async function rollback(deploymentID) {
    busy = true;
    notice = '';
    error = '';
    try {
      await api.rollbackDeployment(id, deploymentID);
      notice = 'Older release restored.';
      await load();
    } catch (err) {
      error = err?.message ?? 'Could not restore this release';
    } finally {
      busy = false;
    }
  }

  function addEnv() {
    envRows = [...envRows, { key: '', value: '', remove: false, existing: false }];
  }

  function toggleRemove(index) {
    const row = envRows[index];
    if (row.existing) row.remove = !row.remove;
    else envRows = envRows.filter((_, i) => i !== index);
  }

  async function saveEnv() {
    const payload = {};
    for (const row of envRows) {
      const key = row.key.trim();
      if (!key) continue;
      if (!KEY.test(key)) {
        notice = `"${key}" is not a valid variable name. Use letters, numbers and underscores.`;
        return;
      }
      if (row.remove) payload[key] = '';
      else if (row.value !== '') payload[key] = row.value;
    }

    if (!Object.keys(payload).length) {
      notice = 'Nothing to save. Values are hidden — type a new value to change one.';
      return;
    }

    savingEnv = true;
    notice = '';
    error = '';
    try {
      const data = await api.updateEnv(id, payload);
      envRows = rowsFromKeys(data?.env_keys ?? []);
      notice = 'Environment variables saved.';
    } catch (err) {
      error = err?.message ?? 'Could not save environment variables';
    } finally {
      savingEnv = false;
    }
  }

  async function saveDomain() {
    savingDomain = true;
    notice = '';
    domainWarning = '';
    error = '';
    try {
      const data = await api.updateDomain(id, domain.trim());
      project = { ...project, custom_domain: data?.custom_domain ?? '' };
      domain = project.custom_domain;
      notice = project.custom_domain ? 'Custom domain saved.' : 'Custom domain removed.';
      // Cloudflare custom hostnames are best effort; surface when SSL is not automatic.
      domainWarning = data?.warning ?? '';
      loadDomainStatus();
    } catch (err) {
      error = err?.message ?? 'Could not save the custom domain';
    } finally {
      savingDomain = false;
    }
  }

  async function loadDomainStatus() {
    if (!id || checkingDomain) return;
    checkingDomain = true;
    try {
      const data = await api.domainStatus(id);
      domainState = data?.state ?? 'unknown';
      domainStatusMessage = data?.message ?? '';
    } catch {
      domainState = 'unknown';
      domainStatusMessage = 'Could not check the domain status right now.';
    } finally {
      checkingDomain = false;
    }
  }

  $effect(() => {
    if (!project?.custom_domain || domainState === 'active') {
      if (domainStatusTimer) {
        window.clearInterval(domainStatusTimer);
        domainStatusTimer = null;
      }
      return;
    }
    if (!domainStatusTimer) domainStatusTimer = window.setInterval(loadDomainStatus, 8000);
  });

  async function copyLog(event, deploymentId, log) {
    event.preventDefault();
    event.stopPropagation();

    const value = log || 'No log recorded.';
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(value);
      } else {
        const helper = document.createElement('textarea');
        helper.value = value;
        helper.style.position = 'fixed';
        helper.style.opacity = '0';
        document.body.appendChild(helper);
        helper.select();
        document.execCommand('copy');
        helper.remove();
      }
      copiedLog = deploymentId;
      setTimeout(() => {
        if (copiedLog === deploymentId) copiedLog = '';
      }, 1600);
    } catch {
      copiedLog = '';
    }
  }

  async function deleteProject() {
    deleting = true;
    error = '';
    try {
      await api.deleteProject(id);
      showDelete = false;
      await reload({ quiet: true });
      navigate('/launchpad');
    } catch (err) {
      error = err?.message ?? 'Could not delete the project';
      showDelete = false;
    } finally {
      deleting = false;
    }
  }
</script>

<Workspace view="projects" compact={true} crumb={project?.subdomain ?? ''}>
  {#snippet children()}
    {#if loading || !project}
      <section class="splash">
        <span class="loading">···</span>
        <p>loading project details…</p>
      </section>
    {:else}
      <section class="detail-head">
        <div class="head-copy">
          <p class="eyebrow"><span class="pulse" aria-hidden="true"></span>{meta.label} project</p>
          <h1>{project.subdomain}<span class="hl">.</span></h1>
          <p class="lead">
            <a href={url} target="_blank" rel="noreferrer">{url} ↗</a>
          </p>
        </div>
        <div class="head-actions">
          <span class="tag {meta.tone}"><i></i>{meta.label}</span>
          <a class="btn" href={url} target="_blank" rel="noreferrer">Open site ↗</a>
          <button type="button" onclick={redeploy} disabled={busy}>{busy ? 'Working…' : 'Redeploy'}</button>
        </div>
      </section>

      {#if error}
        <p class="message bad">{error}</p>
      {/if}
      {#if domainWarning}
        <p class="message warn">{domainWarning}</p>
      {/if}
      {#if !error && notice}
        <p class="message ok">{notice}</p>
      {/if}

      <section class="panel overview">
        <div><span>Repository</span><b>{repo}</b></div>
        <div><span>Created</span><b>{fmt(project.created_at)}</b></div>
        <div><span>Last deploy</span><b>{fmt(project.updated_at)}</b></div>
        <div>
          <span>Custom domain</span>
          {#if project.custom_domain}
            <a
              class="overview-link"
              href={`https://${project.custom_domain}`}
              target="_blank"
              rel="noreferrer"
              aria-label={`Open ${project.custom_domain} in a new tab`}
            >
              {project.custom_domain} ↗
            </a>
          {:else}
            <b>not set</b>
          {/if}
        </div>
      </section>

      <div class="columns">
        <section class="panel traffic">
          <div class="panel-head">
            <div>
              <h2>Traffic</h2>
              <p>Visits over the last 14 days.</p>
            </div>
            <div class="totals">
              <div><b>{totalViews}</b><span>all time</span></div>
              <div><b>{recentViews}</b><span>14 days</span></div>
            </div>
          </div>

          <div
            class="chart"
            role="img"
            aria-label={`${recentViews} visits in the last 14 days`}
            bind:clientWidth={chartWidth}
          >
            <svg viewBox={`0 0 ${chartW} ${CHART_H}`} aria-hidden="true">
              <line class="gridline" x1="0" y1={CHART_H / 4} x2={chartW} y2={CHART_H / 4} />
              <line class="gridline" x1="0" y1={CHART_H / 2} x2={chartW} y2={CHART_H / 2} />
              <line class="gridline" x1="0" y1={(CHART_H / 4) * 3} x2={chartW} y2={(CHART_H / 4) * 3} />

              <path class="area" d={areaPath} />
              <path class="curve" d={linePath} />
              {#each points as point}
                <circle
                  class="point"
                  class:peak={point.count === maxVisits && maxVisits > 0}
                  cx={point.x}
                  cy={point.y}
                  r={point.count === maxVisits && maxVisits > 0 ? 5.5 : 4}
                >
                  <title>{`${point.day}: ${point.count}`}</title>
                </circle>
              {/each}
            </svg>
            <div class="axis">
              {#each visitDays as day, i}
                <span>{i % 3 === 0 || i === 13 ? day.day.slice(5) : ''}</span>
              {/each}
            </div>
          </div>
          <p class="hint">Last visit: {fmt(project.last_visited_at, true)}</p>
        </section>

        <section class="panel domain">
          <div class="panel-head">
            <div>
              <h2>Custom domain</h2>
              <p>Point your domain here, SSL is handled for you.</p>
            </div>
          </div>
          <div class="domain-row">
            <input bind:value={domain} placeholder="www.example.com" aria-label="Custom domain" />
            <button type="button" class="btn" onclick={saveDomain} disabled={savingDomain}>
              {savingDomain ? 'Saving…' : 'Save domain'}
            </button>
          </div>
          {#if project?.custom_domain}
            <div class="domain-status {DOMAIN_STATE_TONE[domainState] ?? 'red'}">
              <span class="status-dot" aria-hidden="true"></span>
              <div>
                <b>{DOMAIN_STATE_LABEL[domainState] ?? 'Checking…'}</b>
                <span>{domainStatusMessage || 'Checking the domain status…'}</span>
              </div>
              {#if domainState !== 'active'}
                <button
                  type="button"
                  class="domain-refresh"
                  onclick={loadDomainStatus}
                  disabled={checkingDomain}
                  aria-label="Check domain status again"
                >
                  <span aria-hidden="true">↻</span>
                  {checkingDomain ? 'Checking…' : 'Check again'}
                </button>
              {/if}
            </div>
          {/if}
          <p class="hint">
            Add this record at your DNS provider, then save the full hostname above (for example
            <code>www.example.com</code>):
          </p>
          <div class="dns-help" aria-label="DNS record example">
            <div><span>Type</span><code>CNAME</code></div>
            <div><span>Name</span><code>{domainLabel}</code></div>
            <div><span>Target</span><code>{CNAME_TARGET}</code></div>
            <div><span>Proxy</span><code>DNS only</code></div>
          </div>
          <p class="hint">
            Name is the first part of the hostname you enter above — for
            <code>mysite.example.com</code>, use <code>mysite</code>. Use a subdomain; most DNS
            providers do not allow CNAME at the root. Keep the proxy off (grey cloud). SSL can take
            a few minutes to activate. Leave the field blank to remove the domain.
          </p>
        </section>
      </div>

      <section class="panel history">
        <div class="panel-head">
          <div>
            <h2>Deploy history</h2>
            <p>Latest builds and their logs. Older successful releases can be restored.</p>
          </div>
        </div>

        {#if !deployments.length}
          <p class="hint">No deployments recorded yet.</p>
        {:else}
          <ul class="deploys">
            {#each deployments as dep, i}
              <li>
                <div class="dep-main">
                  <span class="dot {DEPLOY_TONE[dep.status] ?? 'blue'}"></span>
                  <div>
                    <b>{dep.status}</b>
                    <span>{fmt(dep.created_at, true)}</span>
                  </div>
                  <code>{dep.release || 'no release'}</code>
                </div>
                {#if dep.status === 'success' && dep.release && i > 0}
                  <div class="dep-side">
                    <button type="button" onclick={() => rollback(dep.id)} disabled={busy}>Restore</button>
                  </div>
                {/if}
                <details class="log" bind:open={dep.open}>
                  <summary>
                    <span>
                      Build log
                      {#if isRunning(dep)}
                        <em class="live-pill">Live</em>
                      {/if}
                    </span>
                  </summary>
                  <div class="log-box">
                    <pre>{dep.log || (isRunning(dep) ? 'Waiting for the first build step…' : 'No log recorded.')}</pre>
                    <div class="log-actions">
                      <button
                        type="button"
                        class="copy"
                        aria-label={copiedLog === dep.id ? 'Copied' : 'Copy build log'}
                        title={copiedLog === dep.id ? 'Copied' : 'Copy build log'}
                        onclick={(event) => copyLog(event, dep.id, dep.log)}
                      >
                        {#if copiedLog === dep.id}
                          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <path d="M20 6 9 17l-5-5" />
                          </svg>
                        {:else}
                          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                            <rect x="9" y="9" width="11" height="11" rx="2" />
                            <path d="M5 15V5a2 2 0 0 1 2-2h10" />
                          </svg>
                        {/if}
                      </button>
                    </div>
                  </div>
                </details>
              </li>
            {/each}
          </ul>
        {/if}
      </section>

      <section class="panel env">
        <div class="panel-head">
          <div>
            <h2>Environment</h2>
            <p>Values are write-only. Leave a value blank to keep what is stored.</p>
          </div>
        </div>

        <div class="env-rows">
          {#each envRows as row, i}
            <div class="env-row" class:removed={row.remove}>
              <input bind:value={row.key} placeholder="API_KEY" disabled={row.existing} aria-label="Variable name" />
              <input
                bind:value={row.value}
                type="password"
                placeholder={row.remove ? 'will be removed' : '••••••'}
                disabled={row.remove}
                aria-label="Variable value"
              />
              <button type="button" onclick={() => toggleRemove(i)}>
                {row.existing ? (row.remove ? 'Undo' : 'Remove') : 'Cancel'}
              </button>
            </div>
          {/each}
        </div>

        <div class="env-actions">
          <button type="button" onclick={addEnv}>+ Add variable</button>
          <button type="button" class="btn" onclick={saveEnv} disabled={savingEnv}>
            {savingEnv ? 'Saving…' : 'Save variables'}
          </button>
        </div>
      </section>

      <section class="danger-zone">
        <div>
          <h2>Delete project</h2>
          <p>The site, deployment history, and environment variables will be removed.</p>
        </div>
        <button type="button" class="danger" onclick={() => (showDelete = true)}>
          Delete project
        </button>
      </section>
    {/if}
  {/snippet}
</Workspace>

<ConfirmDialog
  open={showDelete}
  title="Delete project?"
  message={`Delete ${project?.subdomain ?? 'this project'}.only4.fun? This cannot be undone.`}
  confirmLabel="Delete project"
  busy={deleting}
  onConfirm={deleteProject}
  onCancel={() => (showDelete = false)}
/>

<style>
  .splash {
    display: grid;
    place-items: center;
    gap: 16px;
    padding: 120px 0;
  }

  .loading {
    color: var(--blue);
    font: 700 42px/1 'DM Mono', monospace;
    letter-spacing: 0.2em;
  }

  .splash p {
    color: var(--muted);
    font: 500 15px/1 'DM Mono', monospace;
  }

  .detail-head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 28px;
    flex-wrap: wrap;
    padding-bottom: 30px;
  }

  .eyebrow {
    display: inline-flex;
    align-items: center;
    gap: 10px;
    margin: 0 0 16px;
    padding: 8px 16px 8px 14px;
    border: 1px solid var(--line);
    border-radius: 999px;
    font: 500 13px/1 'DM Mono', monospace;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--muted);
  }

  .pulse {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--green);
  }

  h1 {
    margin: 0;
    font-size: clamp(46px, 7vw, 88px);
    line-height: 0.95;
    letter-spacing: -0.045em;
    font-weight: 800;
    word-break: break-word;
  }

  .hl {
    color: var(--yellow);
  }

  .lead {
    margin: 16px 0 0;
    font-size: 19px;
  }

  .lead a {
    color: var(--muted);
    text-decoration: underline;
    text-decoration-thickness: 2px;
    text-underline-offset: 0.18em;
    word-break: break-all;
  }

  .lead a:hover {
    color: var(--ink);
  }

  .head-actions {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  .tag {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 9px 15px;
    border: 1px solid var(--line);
    border-radius: 999px;
    font: 500 13px/1 'DM Mono', monospace;
    text-transform: uppercase;
    color: var(--muted);
  }

  .tag i,
  .dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--blue);
  }

  .tag.green i,
  .dot.green { background: var(--green); }
  .tag.yellow i,
  .dot.yellow { background: var(--yellow); }
  .tag.red i,
  .dot.red { background: var(--red); }

  .btn {
    margin: 0;
    padding: 13px 22px;
    font-size: 16px;
  }

  button {
    padding: 12px 18px;
    border: 2px solid var(--line);
    border-radius: 999px;
    background: transparent;
    color: var(--ink);
    font: 600 15px/1 'Gabarito', sans-serif;
    cursor: pointer;
  }

  button:hover:not(:disabled) {
    border-color: var(--ink);
  }

  button:disabled {
    opacity: 0.6;
    cursor: progress;
  }

  .message {
    margin: 0 0 24px;
    padding: 4px 0 4px 14px;
    border-left: 3px solid currentColor;
    font-size: 16px;
  }

  .message.ok { color: var(--green); }
  .message.warn { color: #9a6700; }
  .message.bad { color: var(--red); }

  .panel {
    padding: 28px 0 0;
    border-top: 1px solid var(--line);
    margin-bottom: 30px;
  }

  .overview {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 20px;
    padding: 24px 0;
    border-bottom: 1px solid var(--line);
    margin-bottom: 30px;
  }

  .overview div {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 0;
  }

  .overview span,
  .totals span {
    color: var(--muted);
    font: 500 12px/1 'DM Mono', monospace;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .overview b,
  .overview-link {
    font-size: 16px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .overview-link {
    color: var(--ink);
    font-weight: 700;
    text-decoration: underline;
    text-decoration-thickness: 2px;
    text-underline-offset: 0.18em;
  }

  .overview-link:hover {
    color: var(--blue);
  }

  .columns {
    display: grid;
    grid-template-columns: minmax(0, 1.3fr) minmax(280px, 0.7fr);
    gap: 28px;
  }

  .columns .traffic {
    padding-right: 28px;
  }

  .columns .domain {
    border-left: 1px solid var(--line);
    padding-left: 28px;
  }

  .panel-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    flex-wrap: wrap;
    margin-bottom: 22px;
  }

  h2 {
    margin: 0 0 6px;
    font-size: clamp(26px, 3vw, 38px);
    line-height: 1;
    letter-spacing: -0.035em;
    font-weight: 800;
  }

  .panel-head p {
    margin: 0;
    color: var(--muted);
    font-size: 15px;
  }

  .domain .panel-head p {
    font-size: 16px;
  }

  .totals {
    display: flex;
    gap: 18px;
  }

  .totals div {
    display: flex;
    flex-direction: column;
    gap: 5px;
    align-items: flex-end;
  }

  .totals b {
    font-size: 26px;
    letter-spacing: -0.04em;
  }

  .chart {
    border-bottom: 2px solid var(--line);
  }

  .chart svg {
    display: block;
    width: 100%;
    height: 180px;
    overflow: visible;
  }

  .gridline {
    stroke: var(--line);
    stroke-width: 1;
    stroke-dasharray: 4 6;
  }

  .area {
    fill: var(--blue);
    fill-opacity: 0.12;
  }

  .curve {
    fill: none;
    stroke: var(--blue);
    stroke-width: 4;
    stroke-linecap: round;
    stroke-linejoin: round;
  }

  .point {
    fill: var(--bg);
    stroke: var(--blue);
    stroke-width: 2.5;
  }

  .point.peak {
    fill: var(--yellow);
    stroke: var(--yellow);
  }

  .axis {
    display: flex;
    gap: 0;
    padding-top: 10px;
  }

  .axis span {
    flex: 1;
    text-align: center;
    color: var(--muted);
    font: 500 10px/1 'DM Mono', monospace;
    white-space: nowrap;
  }

  .hint {
    margin: 18px 0 0;
    color: var(--muted);
    font-family: inherit;
    font-size: 16px;
    line-height: 1.5;
  }

  .hint code {
    font-family: inherit;
    font-weight: 600;
    color: var(--ink);
  }

  .dns-help {
    margin-top: 12px;
    border: 1px solid var(--line);
    border-radius: 12px;
    overflow: hidden;
    font-family: inherit;
  }

  .dns-help > div {
    display: grid;
    grid-template-columns: 62px minmax(0, 1fr);
    gap: 10px;
    align-items: center;
    padding: 8px 12px;
    font-size: 16px;
    line-height: 1.5;
  }

  .dns-help > div + div {
    border-top: 1px solid var(--line);
  }

  .dns-help span {
    color: var(--muted);
  }

  .dns-help code {
    font-family: inherit;
    font-weight: 600;
    overflow-wrap: anywhere;
  }

  .domain-status {
    display: flex;
    align-items: flex-start;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 14px;
    padding: 10px 12px;
    border: 1px solid var(--line);
    border-left: 3px solid var(--muted);
    border-radius: 10px;
    font-family: inherit;
  }

  .domain-status > div {
    flex: 1;
    min-width: 0;
  }

  .domain-refresh {
    margin-left: auto;
    padding: 8px 12px;
    font-size: 15px;
    white-space: nowrap;
  }

  .domain-refresh span {
    font-size: 17px;
    line-height: 1;
  }

  .domain-status.green {
    border-left-color: var(--green);
  }

  .domain-status.yellow {
    border-left-color: var(--yellow);
  }

  .domain-status.red {
    border-left-color: var(--red);
  }

  .status-dot {
    width: 8px;
    height: 8px;
    margin-top: 5px;
    border-radius: 50%;
    background: var(--muted);
    flex: 0 0 auto;
  }

  .green .status-dot {
    background: var(--green);
  }

  .yellow .status-dot {
    background: var(--yellow);
  }

  .red .status-dot {
    background: var(--red);
  }

  .domain-status b {
    display: block;
    font-size: 16px;
    font-weight: 700;
  }

  .domain-status div span {
    display: block;
    margin-top: 3px;
    color: var(--muted);
    font-size: 16px;
    line-height: 1.5;
  }

  .domain-row {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }

  .domain-row input,
  .env-row input {
    flex: 1;
    min-width: 0;
    min-height: 54px;
    padding: 0 16px;
    border: 2px solid var(--line);
    border-radius: 14px;
    outline: 0;
    background: var(--bg);
    color: var(--ink);
    font: 500 17px/1 'DM Mono', monospace;
  }

  .domain-row input:focus,
  .env-row input:focus {
    border-color: var(--blue);
  }

  .deploys {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .deploys li {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    flex-wrap: wrap;
    padding: 16px 0;
    border-bottom: 1px solid var(--line);
  }

  .deploys li:last-child {
    border-bottom: 0;
  }

  .dep-main,
  .dep-side {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
  }

  .dep-main div {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .dep-main b {
    font-size: 16px;
    text-transform: capitalize;
  }

  .dep-main div span {
    color: var(--muted);
    font-size: 13px;
  }

  .dep-main code {
    padding: 5px 10px;
    border-radius: 8px;
    background: var(--soft);
    color: var(--muted);
    font: 500 15px/1 'Gabarito', sans-serif;
  }

  details.log {
    flex: 1 1 100%;
    color: var(--muted);
    font: 500 16px/1.5 'Gabarito', sans-serif;
  }

  summary {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    cursor: pointer;
  }

  summary > span {
    display: inline-flex;
    align-items: center;
    gap: 10px;
  }

  .live-pill {
    display: inline-flex;
    align-items: center;
    padding: 3px 9px;
    border-radius: 999px;
    background: var(--soft);
    color: var(--blue);
    font: 700 12px/1 'Gabarito', sans-serif;
    font-style: normal;
    letter-spacing: 0.02em;
    text-transform: uppercase;
  }

  .log-actions {
    display: flex;
    justify-content: flex-end;
    margin-top: 10px;
  }

  button.copy {
    display: grid;
    place-items: center;
    width: 34px;
    height: 34px;
    padding: 0;
    border-radius: 50%;
  }

  button.copy svg {
    width: 17px;
    height: 17px;
  }

  pre {
    width: 100%;
    min-height: 240px;
    max-height: 420px;
    overflow: auto;
    margin: 12px 0 0;
    padding: 18px 20px;
    border-radius: 12px;
    background: var(--soft);
    color: var(--ink);
    font: 400 16px/1.6 'Gabarito', sans-serif;
    white-space: pre-wrap;
  }

  .env-rows {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-bottom: 20px;
  }

  .env-actions {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  .env-row {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
  }

  .env-row.removed {
    opacity: 0.5;
  }

  .env-row input:first-child {
    flex: 0 1 240px;
  }

  .danger-zone {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 20px;
    flex-wrap: wrap;
    padding: 28px 0 0;
    border-top: 1px solid var(--line);
  }

  .danger-zone h2 {
    margin-bottom: 8px;
  }

  .danger-zone p {
    margin: 0;
    color: var(--muted);
    font-size: 15px;
  }

  button.danger {
    border-color: var(--red);
    color: var(--red);
  }

  button.danger:hover:not(:disabled) {
    background: var(--red);
    color: #fff;
  }

  @media (max-width: 900px) {
    .columns {
      grid-template-columns: 1fr;
      gap: 0;
    }

    .columns .traffic {
      padding-right: 0;
      padding-bottom: 30px;
    }

    .columns .domain {
      border-left: 0;
      padding-left: 0;
      margin-bottom: 0;
    }

    .head-actions {
      align-items: flex-start;
    }
  }

  @media (max-width: 600px) {
    .env-row input:first-child {
      flex: 1 1 100%;
    }
  }
</style>
