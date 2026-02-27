<!-- SandboxFrame.svelte -->
<script lang="ts">
  import { onDestroy, onMount } from 'svelte'

  import buildHtml from '$shared/components/sandbox/buildHtml'
  import ResizableBox from './ResizableBox.svelte'
  import type { SandboxLog } from './types'

  export let id: string
  export let css = ''
  export let html: string
  export let js: string
  export let resizable = false
  export let className = ''
  export let onUpdateLogs: ((logs: SandboxLog[]) => void) | undefined = undefined
  let iframe: HTMLIFrameElement | null = null
  let logs: SandboxLog[] = []

  const handleSandboxLog = (log: SandboxLog) => {
    logs = [...logs, log]
    onUpdateLogs?.(logs)
  }

  const run = () => {
    logs = []
    onUpdateLogs?.(logs)

    if (!iframe) return
    const content = buildHtml(id, css, html, js)
    iframe.srcdoc = content
  }

  export function refresh() {
    run()
  }

  onMount(() => {
    ;(window as unknown as Record<string, unknown>)[id] = handleSandboxLog
    run()
  })

  onDestroy(() => {
    const win = window as unknown as Record<string, unknown>
    if (win[id] === handleSandboxLog) {
      delete win[id]
    }
  })
</script>

{#if resizable}
  <ResizableBox className={`h-full rounded ${className}`}>
    <iframe bind:this={iframe} title={`sandbox-${id}`} class="w-full h-full border-0 bg-white"></iframe>
  </ResizableBox>
{:else}
  <div class={`h-full rounded ${className}`}>
    <iframe bind:this={iframe} title={`sandbox-${id}`} class="w-full h-full border-0 bg-white"></iframe>
  </div>
{/if}
