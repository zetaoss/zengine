<!-- BoxLang.svelte -->
<script lang="ts">
  import type { Writable } from 'svelte/store'

  import type { Job } from './types'

  export let job: Writable<Job>
  export let seq: number

  $: jobValue = $job
  $: isMain = seq === jobValue.main

  $: logs = jobValue.langOuts?.logs ?? []
  $: images = jobValue.langOuts?.images ?? []

  $: loaded = jobValue.phase != null
  $: hasLogs = logs.length > 0
  $: hasImages = images.length > 0
  $: hideTexLogs = (() => {
    const box = jobValue.boxes[seq]
    return hasLogs && hasImages && !!box?.lang && box.lang.includes('tex')
  })()
</script>

<slot />

{#if loaded && isMain && jobValue.phase === 'succeeded'}
  <div class="rounded-lg bg-(--console-bg) px-4 py-2 text-sm break-all">
    {#if hasLogs || hasImages}
      {#if hasLogs && !hideTexLogs}
        {#each logs as log, index (index)}
          <pre class={`whitespace-pre-wrap ${log?.charAt(0) === '2' ? 'text-red-500' : ''}`}>{log?.slice(1)}</pre>
        {/each}
      {/if}

      {#if hasImages}
        {#each images as img, index (index)}
          <span>
            <img src={`data:image/png;base64,${img}`} alt="" class="bg-white mb-3 mr-3 max-h-96 max-w-96" />
          </span>
        {/each}
      {/if}
    {/if}
  </div>
{/if}
