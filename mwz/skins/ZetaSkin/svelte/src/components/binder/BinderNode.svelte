<script lang="ts">
  import { mdiChevronRight } from '@mdi/js'
  import { onMount, tick } from 'svelte'

  import type { BinderItem } from '$lib/types/binder'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  export let node: BinderItem
  export let depth = 0
  export let wgArticleId: number
  export let binderId: number
  export let onReveal: (() => void) | undefined = undefined

  let expanded = depth === 0
  let rowEl: HTMLElement | null = null

  const key = () => String(node?.id ?? '')
  const storageKey = () => `binder-${binderId}`

  const readMap = () => {
    try {
      const raw = localStorage.getItem(storageKey())
      if (!raw) return {}
      const parsed = JSON.parse(raw)
      return parsed && typeof parsed === 'object' ? (parsed as Record<string, number>) : {}
    } catch {
      return {}
    }
  }

  const writeMap = (map: Record<string, number>) => {
    try {
      localStorage.setItem(storageKey(), JSON.stringify(map))
    } catch {
      // ignore storage failures
    }
  }

  const persist = () => {
    const map = readMap()
    map[key()] = expanded ? 1 : 0
    writeMap(map)
  }

  const toggle = () => {
    expanded = !expanded
    persist()
  }

  const handleReveal = () => {
    if (!expanded) {
      expanded = true
      persist()
    }
    onReveal?.()
  }

  const isCurrent = () => node?.id === wgArticleId
  const isLink = () => !!node?.href
  const hasChildren = () => !!node?.nodes?.length

  onMount(async () => {
    const map = readMap()
    const saved = map[key()]
    if (saved === 0 || saved === 1) expanded = !!saved

    if (isCurrent()) {
      onReveal?.()
      await tick()
      rowEl?.scrollIntoView({ block: 'center' })
    }
  })
</script>

<li class="flex flex-col">
  <div class="row flex items-stretch hover:bg-gray-500/15" bind:this={rowEl}>
    <svelte:element
      this={isLink() ? 'a' : 'div'}
      href={node?.href || undefined}
      class={`flex flex-1 px-2 hover:no-underline ${isCurrent() ? 'font-bold' : ''} ${node?.new ? 'new' : ''}`}
      title={node?.text}
    >
      <span style={`padding-left: ${depth}rem`}>{node?.text}</span>
    </svelte:element>

    {#if hasChildren()}
      <button class="w-6 h-6 grid place-items-center rounded-full hover:bg-gray-500/30" on:click|stopPropagation|preventDefault={toggle}>
        <ZIcon path={mdiChevronRight} class={`transition-transform ${expanded ? 'rotate-90' : ''}`} />
      </button>
    {/if}
  </div>

  {#if hasChildren()}
    <ul class="p-0 m-0" style={`display: ${expanded ? 'block' : 'none'}`}>
      {#each node.nodes as n (n.text)}
        <svelte:self node={n} depth={depth + 1} {wgArticleId} {binderId} onReveal={handleReveal} />
      {/each}
    </ul>
  {/if}
</li>
