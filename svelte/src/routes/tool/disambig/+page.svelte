<script lang="ts">
  import { page } from '$app/stores'

  import ThePagination from '$lib/components/pagination/ThePagination.svelte'
  import type { PaginateData } from '$lib/components/pagination/types'
  import httpy from '$shared/utils/httpy'
  import { displayTitle } from '$shared/utils/mediawiki'
  import { getWikiViewHref } from '$shared/utils/wikiLink'

  interface DisambigNode {
    title: string
    text: string
    description: string
    href: string
    id?: number
    new?: number
  }

  interface DisambigItem {
    id: number
    title: string
    entries: number
    nodes: DisambigNode[]
    created_at: string
    updated_at: string
  }

  interface DisambigResponse {
    data: DisambigItem[]
    total: number
    page: number
    limit: number
    total_pages: number
  }

  type SortField = 'id' | 'title' | 'entries' | 'created' | 'updated'
  type SortOrder = 'asc' | 'desc'

  let items: DisambigItem[] = []
  let totalItems = 0
  let totalPages = 1
  let currentPage = 1
  let loading = true
  let error: string | null = null
  let sortField: SortField = 'title'
  let sortOrder: SortOrder = 'asc'
  let loadRequestId = 0
  const pageSize = 25

  $: fetchPage = Math.max(1, Number($page.url.searchParams.get('page')) || 1)

  function toggleSort(field: SortField): void {
    if (sortField === field) {
      sortOrder = sortOrder === 'asc' ? 'desc' : 'asc'
    } else {
      sortField = field
      sortOrder = field === 'entries' || field === 'created' || field === 'updated' ? 'desc' : 'asc'
    }
  }

  function formatDateTime(dateStr: string): string {
    if (!dateStr) return '-'
    const date = new Date(dateStr)
    if (isNaN(date.getTime())) return dateStr
    const yyyy = date.getFullYear()
    const mm = String(date.getMonth() + 1).padStart(2, '0')
    const dd = String(date.getDate()).padStart(2, '0')
    const hh = String(date.getHours()).padStart(2, '0')
    const min = String(date.getMinutes()).padStart(2, '0')
    return `${yyyy}-${mm}-${dd} ${hh}:${min}`
  }

  async function loadDisambigs(pageNum: number, sort: SortField, order: SortOrder) {
    const requestId = ++loadRequestId
    loading = true
    error = null
    try {
      const url = `/api/disambigs?page=${pageNum}&limit=${pageSize}&sort=${sort}&order=${order}&_ts=${Date.now()}`
      const [resData, err] = await httpy.get<DisambigResponse | DisambigItem[]>(url)
      if (requestId !== loadRequestId) return
      if (err) {
        throw err
      }
      if (resData && !Array.isArray(resData) && typeof resData === 'object' && 'data' in resData) {
        items = resData.data || []
        totalItems = resData.total || 0
        totalPages = resData.total_pages || 1
        currentPage = resData.page || pageNum
      } else if (Array.isArray(resData)) {
        items = resData
        totalItems = resData.length
        totalPages = Math.max(1, Math.ceil(totalItems / pageSize))
        currentPage = pageNum
      }
    } catch (e) {
      if (requestId !== loadRequestId) return
      error = e instanceof Error ? e.message : '동음이의 목록을 불러오지 못했습니다.'
    } finally {
      if (requestId === loadRequestId) {
        loading = false
      }
    }
  }

  $: {
    loadDisambigs(fetchPage, sortField, sortOrder)
  }

  $: startIdx = (currentPage - 1) * pageSize
  $: endIdx = Math.min(totalItems, currentPage * pageSize)

  $: paginateData = {
    current_page: currentPage,
    last_page: totalPages,
    path: '/tool/disambig' as const,
  } satisfies PaginateData
</script>

<div class="p-4 md:p-6">
  <div class="mb-5 flex items-baseline gap-2">
    <h1 class="text-2xl font-bold tracking-tight">동음이의</h1>
    {#if !loading && !error}
      <p class="text-sm text-muted-foreground">{totalItems}개 문서</p>
    {/if}
  </div>

  {#if loading}
    <p class="py-8 text-center text-muted-foreground">로딩 중...</p>
  {:else if error}
    <p class="py-8 text-center text-destructive">오류: {error}</p>
  {:else if items.length === 0}
    <p class="py-8 text-center text-muted-foreground">등록된 동음이의어 문서가 없습니다.</p>
  {:else}
    <div class="w-full overflow-x-auto rounded-2xl border border-border bg-card shadow-sm">
      <table class="w-full text-left text-sm border-collapse">
        <thead class="bg-muted/50 text-xs font-semibold text-muted-foreground uppercase border-b border-border">
          <tr class="select-none">
            <th
              class="px-4 py-3 w-20 text-center cursor-pointer hover:bg-muted/70 transition-colors"
              onclick={() => toggleSort('id')}
            >
              <div class="inline-flex items-center justify-center gap-1">
                <span>ID</span>
                {#if sortField === 'id'}
                  <span class="text-[10px] text-primary">{sortOrder === 'asc' ? '▲' : '▼'}</span>
                {/if}
              </div>
            </th>
            <th
              class="px-4 py-3 min-w-[140px] text-left cursor-pointer hover:bg-muted/70 transition-colors"
              onclick={() => toggleSort('title')}
            >
              <div class="inline-flex items-center gap-1">
                <span>기본 제목</span>
                {#if sortField === 'title'}
                  <span class="text-[10px] text-primary">{sortOrder === 'asc' ? '▲' : '▼'}</span>
                {/if}
              </div>
            </th>
            <th
              class="px-4 py-3 text-center w-24 cursor-pointer hover:bg-muted/70 transition-colors"
              onclick={() => toggleSort('entries')}
            >
              <div class="inline-flex items-center justify-center gap-1">
                <span>문서 수</span>
                {#if sortField === 'entries'}
                  <span class="text-[10px] text-primary">{sortOrder === 'asc' ? '▲' : '▼'}</span>
                {/if}
              </div>
            </th>
            <th class="px-4 py-3 min-w-[280px] text-left">문서</th>
            <th
              class="px-4 py-3 text-center w-36 cursor-pointer hover:bg-muted/70 transition-colors"
              onclick={() => toggleSort('created')}
            >
              <div class="inline-flex items-center justify-center gap-1">
                <span>생성일시</span>
                {#if sortField === 'created'}
                  <span class="text-[10px] text-primary">{sortOrder === 'asc' ? '▲' : '▼'}</span>
                {/if}
              </div>
            </th>
            <th
              class="px-4 py-3 text-center w-36 cursor-pointer hover:bg-muted/70 transition-colors"
              onclick={() => toggleSort('updated')}
            >
              <div class="inline-flex items-center justify-center gap-1">
                <span>수정일시</span>
                {#if sortField === 'updated'}
                  <span class="text-[10px] text-primary">{sortOrder === 'asc' ? '▲' : '▼'}</span>
                {/if}
              </div>
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-border">
          {#each items as item (item.id)}
            <tr class="hover:bg-muted/30 transition-colors">
              <td class="px-4 py-3 text-center text-xs font-mono text-muted-foreground">
                #{item.id}
              </td>
              <td class="px-4 py-3 font-semibold align-top">
                {displayTitle(item.title)}
              </td>
              <td class="px-4 py-3 text-center align-top font-normal tabular-nums text-xs text-foreground">
                {item.entries}
              </td>
              <td class="px-4 py-3 align-top">
                {#if item.nodes && item.nodes.length > 0}
                  <div class="flex flex-wrap items-center gap-y-1 my-0.5">
                    {#each item.nodes as node, i (node.title + (node.href || ''))}
                      {#if i > 0}
                        <span class="inline-block px-2 text-muted-foreground/40 select-none text-xs">|</span>
                      {/if}
                      <a
                        href={getWikiViewHref(node.title)}
                        rel="external"
                        data-sveltekit-reload
                        title={node.description ? `${displayTitle(node.title)} (${node.description})` : displayTitle(node.title)}
                        class={node.new || !node.id ? 'text-destructive font-medium' : ''}
                      >
                        {displayTitle(node.title)}
                      </a>
                    {/each}
                  </div>
                {:else}
                  <span class="text-xs text-muted-foreground">(항목 없음)</span>
                {/if}
              </td>
              <td class="px-4 py-3 text-center text-xs text-muted-foreground tabular-nums align-top">
                {formatDateTime(item.created_at)}
              </td>
              <td class="px-4 py-3 text-center text-xs text-muted-foreground tabular-nums align-top">
                {formatDateTime(item.updated_at)}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <!-- Reused Shared Pagination Component -->
    {#if totalPages > 1}
      <div class="mt-6 flex flex-col items-center justify-center gap-3">
        <ThePagination {paginateData} />
        <p class="text-xs text-muted-foreground">
          전체 {totalItems}개 중 {totalItems > 0 ? startIdx + 1 : 0}–{endIdx} 표시
        </p>
      </div>
    {/if}
  {/if}
</div>
