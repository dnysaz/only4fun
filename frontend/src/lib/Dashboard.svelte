<script>
  import Workspace from './Workspace.svelte';
  import Illustration from './Illustration.svelte';
  import ProjectCard from './ProjectCard.svelte';
  import { navigate } from './router.svelte.js';
  import { workspace, reload } from './workspace.svelte.js';
</script>

<Workspace view="projects">
  {#snippet children()}
    <section class="projects">
      <div class="section-head">
        <h2>Your projects</h2>
        <div class="tools">
          <button
            class="refresh"
            type="button"
            onclick={() => reload({ quiet: true })}
            disabled={workspace.refreshing}
          >
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path d="M20 11a8 8 0 1 0-2.3 5.7M20 4v7h-7" />
            </svg>
            {workspace.refreshing ? 'Refreshing…' : 'Refresh'}
          </button>
        </div>
      </div>

      {#if workspace.error}
        <div class="notice">
          <div>
            <b>Something went wrong.</b>
            <p>{workspace.error}</p>
          </div>
          <button class="btn" type="button" onclick={() => reload()}>Try again</button>
        </div>
      {:else if !workspace.projects.length}
        <div class="empty">
          <div class="empty-art" aria-hidden="true"><Illustration /></div>
          <div>
            <h3>Nothing launched yet.</h3>
            <p>Your first site is one paste away. Open <a href="/deploy" onclick={(e) => { e.preventDefault(); navigate('/deploy'); }}>Deploy new</a> and hit <b>Start deploy</b>.</p>
          </div>
        </div>
      {:else}
        <div class="list">
          {#each workspace.projects as project (project.id)}
            <ProjectCard {project} onChanged={() => reload({ quiet: true })} />
          {/each}
        </div>
      {/if}
    </section>
  {/snippet}
</Workspace>

<style>
  .projects {
    padding-top: 4px;
  }

  .section-head {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    gap: 18px;
    margin-bottom: 22px;
  }

  h2 {
    margin: 0;
    font-size: clamp(38px, 5.2vw, 68px);
    line-height: 1;
    letter-spacing: -0.04em;
    font-weight: 800;
  }

  .tools {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .refresh {
    display: inline-flex;
    align-items: center;
    gap: 9px;
    background: transparent;
    border: 1px solid var(--line);
    color: var(--ink);
    border-radius: 999px;
    padding: 11px 20px;
    font: 600 15px/1 'Gabarito', sans-serif;
    cursor: pointer;
  }

  .refresh:hover:not(:disabled) {
    background: var(--soft);
  }

  .refresh:disabled {
    opacity: 0.6;
    cursor: progress;
  }

  .notice .btn {
    margin: 0;
    padding: 13px 24px;
    font-size: 16px;
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .empty {
    display: flex;
    align-items: center;
    gap: clamp(24px, 4vw, 52px);
    flex-wrap: wrap;
    padding: clamp(26px, 3vw, 42px);
    border: 2px dashed var(--line);
    border-radius: 32px;
    background: var(--soft);
  }

  .empty-art {
    width: min(280px, 100%);
    flex: 0 1 280px;
  }

  .empty-art :global(.art) {
    width: 100%;
  }

  h3 {
    margin: 0 0 10px;
    font-size: clamp(30px, 3.4vw, 44px);
    line-height: 1;
    letter-spacing: -0.035em;
  }

  .empty p {
    margin: 0;
    max-width: 40ch;
    color: var(--muted);
    font-size: 19px;
  }

  .empty a {
    color: var(--blue);
    text-underline-offset: 0.18em;
  }

  .notice {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 18px;
    flex-wrap: wrap;
    padding: 22px 26px;
    border: 2px solid var(--red);
    border-radius: 24px;
  }

  .notice b {
    font-size: 18px;
  }

  .notice p {
    margin: 6px 0 0;
    color: var(--red);
  }

  @media (max-width: 900px) {
    .section-head {
      align-items: flex-start;
      flex-direction: column;
    }
  }
</style>
