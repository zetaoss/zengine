<script lang="ts">
  import { mdiChevronLeft, mdiChevronRight, mdiChevronDoubleLeft, mdiChevronDoubleRight } from '@mdi/js'

  import ZIcon from '$shared/ui/ZIcon.svelte'

  import type { PaginateData } from './types'

  export let paginateData: PaginateData
  export let blockSize = 10

  $: current = paginateData.current_page
  $: last = Math.max(1, paginateData.last_page)

  $: block = (() => {
    const size = blockSize
    const blockStart = Math.floor((current - 1) / size) * size + 1
    const blockEnd = Math.min(blockStart + size - 1, last)

    const pages = Array.from({ length: blockEnd - blockStart + 1 }, (_, i) => blockStart + i)

    return {
      blockStart,
      blockEnd,
      pages,
      hasPrevBlock: blockStart > 1,
      hasPrevPage: current > 1,
      hasNextPage: current < last,
      hasNextBlock: blockEnd < last,
    }
  })()

  const baseClass =
    'inline-flex items-center justify-center min-w-8 h-8 px-2 rounded-lg border border-border transition text-foreground ' +
    'hover:no-underline hover:bg-muted whitespace-nowrap text-xs'
  const activeClass = 'font-semibold bg-primary text-primary-foreground shadow-sm border-primary'
  const disabledClass = 'opacity-40 cursor-default pointer-events-none'

  function getPageHref(page: number) {
    const base = paginateData.path
    if (page <= 1) return base
    return `${base}?page=${page}`
  }
</script>

<nav class="flex w-full justify-center">
  <div class="max-w-full overflow-x-auto px-2">
    <div class="inline-flex items-center gap-1.5 whitespace-nowrap select-none">
      <!-- 10 pages back -->
      <a
        href={getPageHref(block.blockStart - 1)}
        class={`${baseClass} ${block.hasPrevBlock ? '' : disabledClass}`}
        aria-label="Previous 10 pages"
        title="이전 10개 페이지"
      >
        <ZIcon path={mdiChevronDoubleLeft} />
      </a>

      <!-- 1 page back -->
      <a
        href={getPageHref(current - 1)}
        class={`${baseClass} ${block.hasPrevPage ? '' : disabledClass}`}
        aria-label="Previous page"
        title="이전 페이지"
      >
        <ZIcon path={mdiChevronLeft} />
      </a>

      {#each block.pages as page (page)}
        <a
          href={getPageHref(page)}
          class={`${baseClass} ${page === current ? activeClass : ''}`}
          aria-current={page === current ? 'page' : undefined}
        >
          {page}
        </a>
      {/each}

      <!-- 1 page forward -->
      <a
        href={getPageHref(current + 1)}
        class={`${baseClass} ${block.hasNextPage ? '' : disabledClass}`}
        aria-label="Next page"
        title="다음 페이지"
      >
        <ZIcon path={mdiChevronRight} />
      </a>

      <!-- 10 pages forward -->
      <a
        href={getPageHref(block.blockEnd + 1)}
        class={`${baseClass} ${block.hasNextBlock ? '' : disabledClass}`}
        aria-label="Next 10 pages"
        title="다음 10개 페이지"
      >
        <ZIcon path={mdiChevronDoubleRight} />
      </a>
    </div>
  </div>
</nav>
