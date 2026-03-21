<script lang="ts">
  import { mdiMinus, mdiPlus } from '@mdi/js'
  import { onMount, tick } from 'svelte'

  import type { BinderItem } from '$lib/types/binder'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  export let node: BinderItem
  export let depth = 0
  export let wgArticleId: number
  export let binderId: number
  export let pathKey = ''
  export let onReveal: (() => void) | undefined = undefined

  let expanded = depth === 0
  let rowEl: HTMLElement | null = null

  const key = () => pathKey
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
  const isInlineToggle = () => !isLink() && hasChildren()
  const isSplitToggle = () => hasChildren() && !isInlineToggle()

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
  <div class={`row flex items-stretch gap-px rounded-lg ${isSplitToggle() ? 'row-split' : 'row-unified'}`} bind:this={rowEl}>
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <svelte:element
      this={isLink() ? 'a' : isInlineToggle() ? 'button' : 'div'}
      href={node?.href || undefined}
      type={isInlineToggle() ? 'button' : undefined}
      class={`binder-link flex min-w-0 flex-1 items-center px-2 py-1 ${
        isInlineToggle() ? 'cursor-pointer justify-between pr-0 text-left rounded-lg' : isSplitToggle() ? 'rounded-lg' : 'rounded-lg'
      } ${isCurrent() ? 'font-bold' : ''} ${node?.new ? 'z-link-new' : isLink() ? 'z-link' : ''}`}
      title={node?.text}
      aria-expanded={isInlineToggle() ? expanded : undefined}
      on:click={isInlineToggle() ? toggle : undefined}
    >
      <span class="binder-label" style={`padding-left: ${depth}rem`}>{node?.text}</span>

      {#if isInlineToggle()}
        <span class="binder-toggle-slot" aria-hidden="true">
          <span class={`binder-expander transition-transform duration-150 ${expanded ? 'rotate-180' : 'rotate-0'}`}>
            <ZIcon path={expanded ? mdiMinus : mdiPlus} class="h-3.5 w-3.5" />
          </span>
        </span>
      {/if}
    </svelte:element>

    {#if hasChildren() && !isInlineToggle()}
      <button
        class="binder-toggle grid h-auto w-8 shrink-0 place-items-center rounded-lg cursor-pointer"
        aria-label={expanded ? 'Collapse binder section' : 'Expand binder section'}
        on:click|stopPropagation|preventDefault={toggle}
      >
        <span class={`binder-expander transition-transform duration-150 ${expanded ? 'rotate-180' : 'rotate-0'}`} aria-hidden="true">
          <ZIcon path={expanded ? mdiMinus : mdiPlus} class="h-3.5 w-3.5" />
        </span>
      </button>
    {/if}
  </div>

  {#if hasChildren()}
    <ul class="p-0 m-0" style={`display: ${expanded ? 'block' : 'none'}`}>
      {#each node.nodes as n, index (`${pathKey}.${index}`)}
        <svelte:self node={n} depth={depth + 1} {wgArticleId} {binderId} pathKey={`${pathKey}.${index}`} onReveal={handleReveal} />
      {/each}
    </ul>
  {/if}
</li>

<style>
  .row {
    transition: background-color 120ms ease;
  }

  .row-split:hover {
    background-color: var(--x-gray-50);
  }

  .row-unified:hover {
    background-color: var(--x-gray-100);
  }

  .binder-link:hover {
    background-color: var(--x-gray-100);
    text-decoration: none;
  }

  .binder-toggle:hover {
    background-color: var(--x-gray-200);
  }

  .binder-toggle-slot {
    display: grid;
    width: 2rem;
    flex: 0 0 auto;
    place-items: center;
  }

  .binder-expander {
    display: inline-grid;
    width: 1.35rem;
    height: 1.35rem;
    flex: 0 0 auto;
    color: var(--x-gray-500);
    font-size: 0.95rem;
    font-weight: 700;
    line-height: 1;
    place-items: center;
  }

  .binder-link:hover .binder-expander,
  .binder-toggle:hover .binder-expander,
  .row:hover .binder-expander {
    color: var(--x-gray-700);
  }
</style>
