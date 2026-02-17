<script lang="ts">
  import type { Section } from '$lib/types/toc'

  import TocNode from './TocNode.svelte'

  export let section: Section
  export let targetIds: string[] = []
  export let depth = 0
  export let showRail = true
  export let onNavigate: (id: string) => void = () => {}

  $: anchor = section?.anchor ?? ''
  $: children = section?.['array-sections'] ?? []
  $: label = (section?.line ?? '').replace(/<\/?[^>]+>/gi, ' ').trim()
  $: number = section?.number ?? ''
  $: isInView = !!anchor && targetIds.includes(anchor)

  const onClick = (e: MouseEvent) => {
    e.preventDefault()
    if (!anchor) return
    if (onNavigate) onNavigate(anchor)
  }
</script>

<a
  href={anchor ? `#${anchor}` : '#'}
  class={`flex w-full items-start gap-1 z-text2 hover:no-underline hover:text-(--link) ${showRail ? 'border-l-2' : ''}`}
  style={`padding-left: calc((${depth} + 1) * 0.75rem); ${showRail ? `border-color: ${isInView ? '#999' : '#9993'};` : ''}`}
  aria-current={isInView ? 'location' : undefined}
  on:click={onClick}
>
  <span class="shrink-0 z-text4">
    <span>{number}</span>
    {#if depth === 0}
      <span>.</span>
    {/if}
  </span>
  <span class="flex-1">{label}</span>
</a>

{#if children?.length > 0}
  <ul class="p-0 list-none" role="list">
    {#each children as s (s.index ?? s.anchor)}
      <li class="m-0">
        <TocNode section={s} {targetIds} depth={depth + 1} {showRail} {onNavigate} />
      </li>
    {/each}
  </ul>
{/if}
