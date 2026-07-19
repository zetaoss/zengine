<!-- BoxLang.svelte -->
<script lang="ts">
  import { mdiAlert } from '@mdi/js'
  import type { Writable } from 'svelte/store'

  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'

  import type { Job } from './types'

  export let job: Writable<Job>
  export let seq: number
  export let wrapped = true

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

  $: lowerPhase = jobValue.phase?.toLowerCase()

  const containerClass = 'rounded-b-lg border-t px-4 py-2 text-sm'
</script>

<slot />

{#if loaded && isMain && (lowerPhase === 'pending' || lowerPhase === 'running')}
  <div class="{containerClass} flex items-center">
    {#if lowerPhase === 'running'}
      <ZSpinner size="0.875rem" extraClass="mr-0!" />
    {:else if lowerPhase === 'pending'}
      <ZSpinner size="0.875rem" extraClass="mr-0! [animation-direction:reverse]" />
    {/if}
    <span class="text-muted-foreground text-xs">{lowerPhase}...</span>
  </div>
{/if}

{#if loaded && isMain && (lowerPhase === 'succeeded' || lowerPhase === 'failed')}
  <div class="{containerClass} max-h-240 overflow-y-auto" class:break-all={wrapped} class:overflow-x-auto={!wrapped}>
    {#if lowerPhase === 'failed'}
      <div class="flex items-center text-xs text-a-red-500 font-medium mb-2">
        <ZIcon size={14} path={mdiAlert} />
      </div>
    {/if}

    {#if hasLogs || hasImages}
      {#if hasLogs && !hideTexLogs}
        {#each logs as log, index (index)}
          <pre
            class={`${wrapped ? 'whitespace-pre-wrap' : 'whitespace-pre'} ${log?.charAt(0) === '2' ? 'text-a-red-500' : ''}`}>{log?.slice(
              1,
            )}</pre>
        {/each}
      {/if}

      {#if hasImages}
        {#each images as img, index (index)}
          <span>
            <img src={`data:image/png;base64,${img}`} alt="" class="border rounded bg-white m-2 ml-0 max-h-96 max-w-96" />
          </span>
        {/each}
      {/if}
    {/if}
  </div>
{/if}
