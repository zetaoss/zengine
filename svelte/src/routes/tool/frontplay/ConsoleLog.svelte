<script lang="ts">
  import { type ConsoleLogLevel, getConsoleEmoji } from '$shared/utils/console'

  import ConsoleArg from './ConsoleArg.svelte'

  export let level: ConsoleLogLevel
  export let args: unknown[]

  $: emoji = getConsoleEmoji(level)
</script>

<div class={`border-t px-2 py-0.5 ${level}`}>
  <div class="flex gap-2">
    <span class="w-4">{emoji}</span>
    {#each args as arg, i (i)}
      <ConsoleArg value={arg} />
    {/each}
  </div>
</div>
