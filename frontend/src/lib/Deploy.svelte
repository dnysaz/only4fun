<script>
  import Workspace from './Workspace.svelte';
  import CreateProject from './CreateProject.svelte';
  import { getQuery } from './router.svelte.js';
  import { reload } from './workspace.svelte.js';

  let preset = $state(getQuery().get('claim') ?? '');

  async function onCreated() {
    preset = '';
    await reload({ quiet: true });
  }
</script>

<Workspace view="deploy">
  {#snippet children()}
    <CreateProject {preset} {onCreated} />
  {/snippet}
</Workspace>
