<script lang="ts">
  import { onDestroy, onMount } from 'svelte'

  import buildHtml from '$shared/components/sandbox/buildHtml'
  import ZButton from '$shared/ui/ZButton.svelte'

  import ConsoleLog from './frontplay/ConsoleLog.svelte'
  import type { LogLevel, SandboxLog } from './frontplay/types'

  interface LogItem extends SandboxLog {
    id: number
  }

  let htmlCode = `<h1>Hello, World!</h1>

<${'script'}>
console.log("Hello HTML");
<${'/script'}>`

  let jsCode = `console.log('Hello JS');`

  let logs: LogItem[] = []
  let logSeed = 0

  let iframeSrcDoc = ''
  const bridgeId = `frontplay_${Math.random().toString(36).slice(2)}`

  function appendLog(level: LogLevel, args: unknown[]) {
    logs = [...logs, { id: (logSeed += 1), level, args }]
  }

  function run() {
    logs = []
    iframeSrcDoc = `${buildHtml(bridgeId, htmlCode, jsCode)}\n<!-- run:${Date.now()} -->`
  }

  if (typeof window !== 'undefined') {
    const bridgeWindow = window as unknown as Record<string, unknown>
    bridgeWindow[bridgeId] = (payload: unknown) => {
      const data = payload as { level?: LogLevel; args?: unknown[] }
      if (!data?.level || !Array.isArray(data.args)) return
      appendLog(data.level, data.args)
    }
  }

  onDestroy(() => {
    if (typeof window !== 'undefined') {
      const bridgeWindow = window as unknown as Record<string, unknown>
      delete bridgeWindow[bridgeId]
    }
  })

  onMount(() => {
    run()
  })
</script>

<div class="grid grid-cols-1 gap-4 p-4 md:grid-cols-2">
  <div>
    <div class="flex items-center py-2">
      <b>Front Play</b>
      <span class="ml-2">
        <ZButton onclick={run}>Run</ZButton>
      </span>
    </div>

    <div class="space-y-2">
      <textarea bind:value={htmlCode} class="h-[35vh] w-full resize-none rounded border p-2 font-mono"></textarea>
      <textarea bind:value={jsCode} class="h-[35vh] w-full resize-none rounded border p-2 font-mono"></textarea>
    </div>
  </div>

  <div class="flex flex-col gap-2">
    <div class="h-[50vh] overflow-hidden rounded border">
      <iframe title="FrontPlay Sandbox" srcdoc={iframeSrcDoc} class="h-full w-full"></iframe>
    </div>

    <div class="flex min-h-0 flex-1 flex-col">
      <header class="bg-slate-400 py-1 text-center font-bold text-white dark:bg-slate-600">Console</header>
      <div class="console h-[30vh] overflow-y-auto bg-(--console-bg) text-sm">
        {#if logs.length === 0}
          <div class="p-2 opacity-60">No logs yet.</div>
        {:else}
          {#each logs as log (log.id)}
            <ConsoleLog level={log.level} args={log.args} />
          {/each}
        {/if}
      </div>
    </div>
  </div>
</div>

<style>
  .console {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
  }

  .console :global(.string) {
    color: var(--console-string);
  }

  .console :global(.boolean),
  .console :global(.number) {
    color: var(--console-number);
  }

  .console :global(.null),
  .console :global(.undefined) {
    color: var(--console-null);
  }

  .console :global(.circular) {
    color: var(--console-circular);
    font-style: italic;
  }

  .console :global(.warn) {
    background-color: var(--console-warn-bg);
  }

  .console :global(.error) {
    background-color: var(--console-error-bg);
  }

  .console :global(.detail) {
    margin: 0.25rem 0;
    border-radius: 0.25rem;
    background-color: var(--console-detail-bg);
    padding: 0.25rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
  }

  .console :global(.arrkey) {
    color: var(--console-arrkey);
  }

  .console :global(.objkey) {
    color: var(--console-objkey);
  }
</style>
