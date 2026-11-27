<script>
  import Icon from './Icon.svelte';
  import { navigate } from './router.svelte.js';

  let { project } = $props();

  const STATUS = {
    deployed: { label: 'Live', tone: 'green', icon: 'world' },
    building: { label: 'Building', tone: 'yellow', icon: 'bolt' },
    pending: { label: 'Queued', tone: 'yellow', icon: 'bolt' },
    failed: { label: 'Failed', tone: 'red', icon: 'shield' },
  };

  let meta = $derived(STATUS[project.status] ?? { label: project.status, tone: 'blue', icon: 'world' });
  let url = $derived(project.deploy_url || `https://${project.subdomain}.only4.fun`);
  let repo = $derived(
    project.repo_url?.startsWith('zip://')
      ? 'Uploaded ZIP'
      : (project.repo_url ?? '').replace(/^https?:\/\/(www\.)?github\.com\//, ''),
  );
  let views = $derived(
    project.views >= 1000 ? `${(project.views / 1000).toFixed(1)}k` : String(project.views ?? 0),
  );
  let updated = $derived(
    project.updated_at
      ? new Date(project.updated_at).toLocaleDateString(undefined, {
          day: 'numeric',
          month: 'short',
        })
      : '—',
  );

  function goDetail(event) {
    event.preventDefault();
    navigate(`/project/${project.id}`);
  }
</script>

<article class="card">
  <a
    class="detail-link"
    href={`/project/${project.id}`}
    onclick={goDetail}
    aria-label={`Open details for ${project.subdomain}`}
  ></a>

  <div class="lead">
    <div class="ic {meta.tone}"><Icon name={meta.icon} /></div>
    <div class="title">
      <h3>{project.subdomain}</h3>
      <code>{project.subdomain}.only4.fun</code>
    </div>
    <span class="tag {meta.tone}"><i></i>{meta.label}</span>
    <span class="more" aria-hidden="true">→</span>
  </div>

  <div class="meta">
    <p class="repo mono" title={project.repo_url}>{repo}</p>
    <div class="facts">
      <span><b>{views}</b> views</span>
      <span>updated <b>{updated}</b></span>
    </div>
  </div>

  <div class="actions">
    <a class="btn open" href={url} target="_blank" rel="noreferrer">
      Open
      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <path d="M7 17L17 7" /><path d="M8 7h9v9" />
      </svg>
    </a>
  </div>
</article>

<style>
  .card {
    position: relative;
    display: grid;
    grid-template-columns: minmax(280px, 1.25fr) minmax(240px, 1fr) auto;
    align-items: center;
    gap: 24px;
    padding: 18px 22px;
    border: 1px solid var(--line);
    border-radius: 16px;
    background: var(--bg);
    transition: background 0.15s;
  }

  .card:hover {
    background: var(--soft);
  }

  .detail-link {
    position: absolute;
    inset: 0;
    z-index: 1;
    border-radius: inherit;
  }

  .detail-link:focus-visible {
    outline: 3px solid var(--blue);
    outline-offset: 3px;
  }

  .lead {
    position: relative;
    z-index: 2;
    pointer-events: none;
    display: flex;
    align-items: center;
    gap: 14px;
    min-width: 0;
  }

  .ic {
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    flex: none;
    border-radius: 50%;
    color: #fff;
  }

  .ic.green { background: var(--green); }
  .ic.red { background: var(--red); }
  .ic.blue { background: var(--blue); }
  .ic.yellow {
    background: var(--yellow);
    /* dark glyph keeps contrast on yellow */
    color: #202124;
  }

  .title {
    min-width: 0;
    flex: 1;
  }

  h3 {
    margin: 0 0 5px;
    font-size: 22px;
    font-weight: 700;
    letter-spacing: -0.025em;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  code {
    display: inline-block;
    max-width: 100%;
    padding: 3px 9px;
    border-radius: 8px;
    background: var(--soft);
    color: var(--muted);
    font: 500 13px/1.3 'DM Mono', monospace;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    vertical-align: middle;
  }

  .tag {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    flex: none;
    padding: 7px 13px;
    border: 1px solid var(--line);
    border-radius: 999px;
    font: 500 12px/1 'DM Mono', monospace;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--muted);
  }

  .tag i {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--blue);
  }

  .tag.green i { background: var(--green); }
  .tag.yellow i { background: var(--yellow); }
  .tag.red i { background: var(--red); }

  .more {
    color: var(--muted);
    font-size: 22px;
    line-height: 1;
  }

  .card:hover .more {
    color: var(--ink);
    transform: translateX(3px);
  }

  .meta {
    position: relative;
    z-index: 2;
    pointer-events: none;
    display: flex;
    flex-direction: column;
    gap: 9px;
    min-width: 0;
  }

  .repo {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0;
    color: var(--muted);
    font-size: 14px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .repo::before {
    content: '';
    width: 8px;
    height: 8px;
    flex: none;
    border-radius: 50%;
    background: var(--line);
  }

  .facts {
    display: flex;
    gap: 20px;
    flex-wrap: wrap;
    color: var(--muted);
    font-size: 15px;
  }

  .facts b {
    color: var(--ink);
    font-weight: 800;
  }

  .actions {
    position: relative;
    z-index: 2;
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: nowrap;
    justify-content: flex-end;
  }

  .open {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    margin: 0;
    padding: 11px 18px;
    font-size: 15px;
  }

  @media (max-width: 1100px) {
    .card {
      grid-template-columns: minmax(0, 1fr) auto;
      gap: 14px 20px;
    }

    .meta {
      grid-column: 1 / -1;
      grid-row: 2;
      flex-direction: row;
      align-items: center;
      justify-content: space-between;
      gap: 16px;
    }

    .actions {
      grid-column: 2;
      grid-row: 1;
    }
  }

  @media (max-width: 700px) {
    .card {
      grid-template-columns: 1fr;
      padding: 20px;
    }

    .meta,
    .actions {
      grid-column: 1;
      grid-row: auto;
    }

    .meta {
      flex-direction: column;
      align-items: flex-start;
      gap: 8px;
    }

    .actions {
      justify-content: flex-start;
      flex-wrap: wrap;
    }
  }
</style>
