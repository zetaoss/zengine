<!-- BoxLang.svelte -->
<script lang="ts">
  import { mdiAlert } from '@mdi/js'
  import type { Writable } from 'svelte/store'

  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import { formatDateTime } from '$shared/utils/time'

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
  $: displayPhase = jobValue.isLoading ? 'loading' : lowerPhase
  $: updatedAtLabel = formatDateTime(jobValue.updatedAt)

  const containerClass = 'rounded-b-lg border-t px-4 py-2 text-sm'
</script>

<slot />

{#if isMain && (displayPhase === 'loading' || displayPhase === 'pending' || displayPhase === 'running')}
  <div class="{containerClass} flex items-center">
    {#if displayPhase === 'loading' || displayPhase === 'running'}
      <ZSpinner size="0.875rem" extraClass="mr-1!" />
    {:else if displayPhase === 'pending'}
      <ZSpinner size="0.875rem" extraClass="mr-1! [animation-direction:reverse]" />
    {/if}
    <span class="text-muted-foreground text-xs">{displayPhase}...</span>
  </div>
{/if}

{#if loaded && isMain && (lowerPhase === 'succeeded' || lowerPhase === 'failed')}
  <div class="{containerClass} max-h-240 overflow-y-auto" class:break-all={wrapped} class:overflow-x-auto={!wrapped}>
    {#if lowerPhase === 'failed'}
      <div class="flex items-center gap-1 text-xs text-a-red-500 font-medium mb-2">
        <ZIcon size={14} path={mdiAlert} />
        {#if updatedAtLabel}
          <time datetime={jobValue.updatedAt ?? undefined} title={jobValue.updatedAt ?? undefined}>
            Failed · {updatedAtLabel}
          </time>
        {:else}
          <span>Failed</span>
        {/if}
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
