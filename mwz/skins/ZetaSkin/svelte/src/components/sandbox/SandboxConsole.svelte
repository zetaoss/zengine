<!-- SandboxConsole.svelte -->
<script lang="ts">
  import ConsoleLog from './ConsoleLog.svelte'
  import type { SandboxLog } from './types'

  export let logs: SandboxLog[]
  export let className = ''
</script>

<div class={`console font-mono text-sm ${className}`}>
  {#each logs as log, i (i)}
    <ConsoleLog {...log} />
  {/each}
</div>

<style>
  :global(.console .string) {
    color: var(--console-string);
  }

  :global(.console .boolean),
  :global(.console .number) {
    color: var(--console-number);
  }

  :global(.console .null),
  :global(.console .undefined) {
    color: var(--console-null);
  }

  :global(.console .circular) {
    color: var(--console-circular);
    font-style: italic;
  }

  :global(.console .warn) {
    background-color: var(--console-warn-bg);
  }

  :global(.console .error) {
    background-color: var(--console-error-bg);
  }

  :global(.console .detail) {
    padding: 0.25rem;
    margin: 0.25rem 0;
    font-size: 0.875rem;
    line-height: 1.25rem;
    border-radius: 0.25rem;
    background-color: var(--console-detail-bg);
  }

  :global(.console .arrkey) {
    color: var(--console-arrkey);
  }

  :global(.console .objkey) {
    color: var(--console-objkey);
  }
</style>
