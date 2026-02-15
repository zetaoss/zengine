<script lang="ts">
  import ConsoleArg from './ConsoleArg.svelte'
  import type { LogLevel } from './types'

  export let level: LogLevel
  export let args: unknown[]

  function getEmoji(name: LogLevel): string {
    if (typeof window === 'undefined') return ''
    const value = getComputedStyle(document.documentElement).getPropertyValue(`--console-emoji-${name}`).trim()
    return value || ''
  }

  $: emoji = getEmoji(level)
</script>

<div class={`border-t px-2 py-0.5 ${level}`}>
  <div class="flex gap-2">
    <span class="w-4">{emoji}</span>
    {#each args as arg, i (i)}
      <ConsoleArg value={arg} />
    {/each}
  </div>
</div>
