<script lang="ts">
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
  $: cssCode = getCode(jobValue, 'css')
  $: htmlCode = getCode(jobValue, 'html')
  $: jsCode = getCode(jobValue, 'javascript')

  $: sandboxId = `sandbox-${jobValue.id}-${seq}`

  let logs: SandboxLog[] = []
  const updateLogs = (nextLogs: SandboxLog[]) => {
    logs = nextLogs
  }
</script>

<slot />

{#if jobValue.main === seq}
  <SandboxFrame
    id={sandboxId}
    css={cssCode}
    html={htmlCode}
    js={jsCode}
    resizable={jobValue.outResize}
    className={`mt-1 h-32 rounded ${!jobValue.boxes.some((b) => ['html', 'css', 'javascript'].includes(b.lang)) ? 'hidden' : ''}`}
    onUpdateLogs={updateLogs}
  />

  {#if logs.length > 0}
    <div class="max-h-40 overflow-y-auto mt-1 bg-(--console-bg)">
      <SandboxConsole {logs} className="rounded-lg" />
    </div>
  {/if}
{/if}
