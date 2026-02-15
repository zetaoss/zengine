<script lang="ts">
  import ConsoleArg from './ConsoleArg.svelte'
  import type { LogLevel } from './types'

  export let level: LogLevel
  export let args: unknown[]

  $: emoji = getComputedStyle(document.documentElement).getPropertyValue(`--console-emoji-${level}`)
</script>

<div class={`px-2 py-0.5 border-t ${level}`}>
  <div class="flex gap-2">
    <span class="w-4">{emoji ?? ''}</span>
    {#each args as arg, i (i)}
      <ConsoleArg value={arg} />
    {/each}
  </div>
</div>
