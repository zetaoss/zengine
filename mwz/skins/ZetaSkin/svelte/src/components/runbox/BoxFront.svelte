<script lang="ts">
  import { afterUpdate } from 'svelte'
  import type { Writable } from 'svelte/store'

  import SandboxConsole from '../sandbox/SandboxConsole.svelte'
  import SandboxFrame from '../sandbox/SandboxFrame.svelte'
  import type { SandboxLog } from '../sandbox/types'
  import type { Job } from './types'

  export let job: Writable<Job>
  export let seq: number

  const getCode = (current: Job, lang: string) =>
    current.boxes
      .filter((b) => b.lang === lang)
      .map((b) => b.text)
      .join('\n')
      .trim()

  $: jobValue = $job
  $: htmlCode = getCode(jobValue, 'html')
  $: jsCode = getCode(jobValue, 'javascript')

  $: sandboxId = `sandbox-${jobValue.id}-${seq}`

  let logs: SandboxLog[] = []
  const updateLogs = (e: CustomEvent<SandboxLog[]>) => {
    logs = e.detail
  }

  let consoleRef: HTMLElement | null = null
  let lastLogCount = 0

  afterUpdate(() => {
    if (!consoleRef) return
    if (logs.length === lastLogCount) return
    lastLogCount = logs.length
    consoleRef.scrollTop = consoleRef.scrollHeight
  })
</script>

<slot />

{#if jobValue.main === seq}
  <SandboxFrame
    id={sandboxId}
    html={htmlCode}
    js={jsCode}
    resizable={jobValue.outResize}
    className={`mt-1 h-32 rounded ${!jobValue.boxes.some((b) => b.lang === 'html') ? 'hidden' : ''}`}
    on:update:logs={updateLogs}
  />

  {#if logs.length > 0}
    <div bind:this={consoleRef} class="max-h-40 overflow-y-auto mt-1 bg-[var(--console-bg)]">
      <SandboxConsole {logs} className="rounded-lg" />
    </div>
  {/if}
{/if}
