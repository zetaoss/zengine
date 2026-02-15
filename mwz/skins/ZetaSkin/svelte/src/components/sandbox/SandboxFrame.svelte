<!-- SandboxFrame.svelte -->
<script lang="ts">
  import { createEventDispatcher, onDestroy, onMount } from 'svelte'

  import buildHtml from './buildHtml'
  import ResizableBox from './ResizableBox.svelte'
  import type { SandboxLog } from './types'

  export let id: string
  export let html: string
  export let js: string
  export let resizable = false
  export let className = ''

  const dispatch = createEventDispatcher<{ 'update:logs': SandboxLog[] }>()

  let iframe: HTMLIFrameElement | null = null
  let logs: SandboxLog[] = []

  const handleSandboxLog = (log: SandboxLog) => {
    logs = [...logs, log]
    dispatch('update:logs', logs)
  }

  const run = () => {
    logs = []
    dispatch('update:logs', logs)

    if (!iframe) return
    const content = buildHtml(id, html, js)
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
