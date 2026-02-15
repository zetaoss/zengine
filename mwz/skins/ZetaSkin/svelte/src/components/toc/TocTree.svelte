<script lang="ts">
  import type { Section } from '$lib/types/toc'
  import { scrollToTop } from '$shared/utils/scroll'

  import TocNode from './TocNode.svelte'

  export let toc: Section
  export let activeIds: string[] = []
  export let headerOffset = 64
  export let showRail = true

  $: sections = toc?.['array-sections'] ?? []

  const scrollToAnchor = (id: string) => {
    const el = document.getElementById(id)
    if (!el) return
    const top = el.getBoundingClientRect().top + window.scrollY - headerOffset
    window.scrollTo({ top, behavior: 'smooth' })
  }

  const onNavigate = (id: string) => {
    if (!id) return
    scrollToAnchor(id)
    history.pushState(null, '', `#${id}`)
  }
</script>

<nav>
  <ul class="text-sm tracking-tight list-none m-0 p-0 z-text2">
    <li class="m-0 mb-2 flex items-center gap-1">
      <button type="button" on:click={scrollToTop}>목차</button>
    </li>
    {#each sections as s (s.index ?? s.anchor)}
      <li class="m-0">
        <TocNode section={s} targetIds={activeIds} depth={0} {showRail} {onNavigate} />
      </li>
    {/each}
  </ul>
</nav>
