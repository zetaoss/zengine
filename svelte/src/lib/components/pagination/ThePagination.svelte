<script lang="ts">
  import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'

  import ZIcon from '$shared/ui/ZIcon.svelte'

  import type { PaginateData } from './types'

  export let paginateData: PaginateData
  export let blockSize = 10

  $: block = (() => {
    const { current_page, last_page } = paginateData
    const size = blockSize

    const blockStart = Math.floor((current_page - 1) / size) * size + 1
    const blockEnd = Math.min(blockStart + size - 1, last_page)

    const pages = Array.from({ length: blockEnd - blockStart + 1 }, (_, i) => blockStart + i)

    return {
      blockStart,
      blockEnd,
      pages,
      hasPrevBlock: blockStart > 1,
      hasNextBlock: blockEnd < last_page,
    }
  })()

  const baseClass =
    'inline-flex items-center justify-center min-w-8 px-2 sm:px-3 py-2 rounded transition z-text ' +
    'hover:no-underline hover:bg-zinc-100 dark:hover:bg-zinc-800 whitespace-nowrap'
  const activeClass = 'font-bold bg-[#8883]'
  const disabledClass = 'opacity-40 cursor-default pointer-events-none'

  function getPageHref(page: number) {
    const base = paginateData.path
    if (page <= 1) return base
    return `${base}?page=${page}`
  }
</script>

<nav class="flex w-full justify-center">
  <div class="max-w-full overflow-x-auto px-2">
    <div class="inline-flex items-center gap-1 whitespace-nowrap">
      <a
        href={getPageHref(block.blockStart - 1)}
        rel="external"
        class={`${baseClass} ${block.hasPrevBlock ? '' : disabledClass}`}
        aria-label="Previous pages"
      >
        <ZIcon path={mdiChevronLeft} />
      </a>

      {#each block.pages as page (page)}
        <a
          href={getPageHref(page)}
          rel="external"
          class={`${baseClass} ${page === paginateData.current_page ? activeClass : ''}`}
          aria-current={page === paginateData.current_page ? 'page' : undefined}
        >
          {page}
        </a>
      {/each}

      <a
        href={getPageHref(block.blockEnd + 1)}
        rel="external"
        class={`${baseClass} ${block.hasNextBlock ? '' : disabledClass}`}
        aria-label="Next pages"
      >
        <ZIcon path={mdiChevronRight} />
      </a>
    </div>
  </div>
</nav>
