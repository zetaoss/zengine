<svelte:options runes={true} />

<script lang="ts">
  import { mdiArrowDown, mdiArrowUp } from '@mdi/js'
  import { onMount } from 'svelte'

  import useAuthStore from '$lib/stores/auth'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZToggle from '$shared/ui/ZToggle.svelte'
  import httpy, { HttpyError } from '$shared/utils/httpy'

  interface MwAllPage {
    pageid: number
    title: string
  }

  interface MwAllPagesResp {
    query?: {
      allpages?: MwAllPage[]
    }
  }

  interface TplRow {
    enabled: boolean
    id: number
    title: string
  }

  interface ArticleTplConfigResp {
    key: string
    value: number[]
  }

  const auth = useAuthStore()
  const { userInfo } = auth

  let loading = $state(true)
  let saving = $state(false)
  let errorMessage = $state('')
  let rows = $state<TplRow[]>([])
  let enabledOrder = $state<number[]>([])
  let isSysop = $derived(($userInfo?.groups ?? []).includes('sysop'))
  let isBusy = $derived(loading || saving)
  let orderedRows = $derived.by(() => {
    const rank: Record<number, number> = {}
    for (const [index, id] of enabledOrder.entries()) {
      rank[id] = index
    }
    const enabledRows = rows
      .filter((row) => row.enabled)
      .sort((a, b) => (rank[a.id] ?? Number.MAX_SAFE_INTEGER) - (rank[b.id] ?? Number.MAX_SAFE_INTEGER))
    const disabledRows = rows.filter((row) => !row.enabled).sort((a, b) => a.id - b.id)
    return [...enabledRows, ...disabledRows]
  })

  function toWikiPath(title: string) {
    return `/wiki/${title.replaceAll(' ', '_')}`
  }

  async function fetchTitles() {
    const [data, err] = await httpy.get<MwAllPagesResp>('/w/api.php', {
      action: 'query',
      format: 'json',
      formatversion: '2',
      list: 'allpages',
      apnamespace: '10',
      apprefix: '새문서틀',
      aplimit: '50',
    })
    if (err) throw err
    return (data?.query?.allpages ?? []).filter((page) => page.title.startsWith('틀:새문서틀'))
  }

  async function fetchStoredConfig() {
    const [data, err] = await httpy.get<ArticleTplConfigResp>('/api/article-tpl')
    if (err) throw err
    const raw = data?.value
    if (!Array.isArray(raw)) return []
    return raw.filter((value): value is number => typeof value === 'number' && Number.isInteger(value) && value > 0)
  }

  async function fetchStoredConfigSafe() {
    try {
      return await fetchStoredConfig()
    } catch (err) {
      console.error(err)
      return []
    }
  }

  async function saveConfig(nextRows: TplRow[], nextEnabledOrder: number[]) {
    const enabledSet = new Set(nextRows.filter((row) => row.enabled).map((row) => row.id))
    const enabled = nextEnabledOrder.filter((id) => enabledSet.has(id))
    const payload = { enabled }

    saving = true
    const [, err] = await httpy.put('/api/article-tpl', payload)
    saving = false
    if (err) throw err
  }

  function saveFailToast(err: unknown) {
    if (err instanceof HttpyError && err.code === 503) {
      showToast('503 저장 실패')
      return
    }
    showToast('저장 실패')
  }

  function enabledIndex(id: number) {
    return enabledOrder.findIndex((value) => value === id)
  }

  async function updateEnabled(id: number, checked: boolean) {
    if (!isSysop) return
    const prevRows = rows
    const prevEnabledOrder = enabledOrder

    rows = rows.map((row) => (row.id === id ? { ...row, enabled: checked } : row))

    if (checked) {
      if (enabledIndex(id) < 0) {
        enabledOrder = [...enabledOrder, id]
      }
    } else {
      enabledOrder = enabledOrder.filter((value) => value !== id)
    }

    try {
      await saveConfig(rows, enabledOrder)
    } catch (err) {
      console.error(err)
      rows = prevRows
      enabledOrder = prevEnabledOrder
      saveFailToast(err)
    }
  }

  async function moveEnabled(id: number, direction: -1 | 1) {
    if (!isSysop) return
    const currentIndex = enabledIndex(id)
    if (currentIndex < 0) return
    const nextIndex = currentIndex + direction
    if (nextIndex < 0 || nextIndex >= enabledOrder.length) return

    const prevEnabledOrder = enabledOrder
    const nextOrder = [...enabledOrder]
    const temp = nextOrder[currentIndex]
    nextOrder[currentIndex] = nextOrder[nextIndex]
    nextOrder[nextIndex] = temp
    enabledOrder = nextOrder

    try {
      await saveConfig(rows, enabledOrder)
    } catch (err) {
      console.error(err)
      enabledOrder = prevEnabledOrder
      saveFailToast(err)
    }
  }

  async function loadItems() {
    loading = true
    errorMessage = ''

    try {
      const items = await fetchTitles()
      const enabled = await fetchStoredConfigSafe()
      const enabledSet = new Set(enabled)
      enabledOrder = [...enabled]
      rows = items.map((item) => ({
        enabled: enabledSet.has(item.pageid),
        id: item.pageid,
        title: item.title,
      }))
    } catch (err) {
      console.error(err)
      errorMessage = '목록을 불러오지 못했습니다. 잠시 후 다시 시도해 주세요.'
      rows = []
    } finally {
      loading = false
    }
  }

  onMount(() => {
    auth.update()
    void loadItems()
  })
</script>

<div class="p-4 md:p-6">
  <h1 class="text-xl font-bold text-slate-800">문서 템플릿</h1>
  <p class="mt-2 text-sm text-slate-600">문서 작성 시 사용하는 템플릿 목록입니다.</p>

  {#if loading}
    <div class="mt-4 rounded border border-slate-200 bg-white p-4 text-sm text-slate-600">불러오는 중...</div>
  {:else if errorMessage}
    <div class="mt-4 rounded border border-red-200 bg-red-50 p-4 text-sm text-red-700">{errorMessage}</div>
  {:else if rows.length === 0}
    <div class="mt-4 rounded border border-slate-200 bg-white p-4 text-sm text-slate-600">등록된 새문서 틀이 없습니다.</div>
  {:else}
    <div class="mt-4 overflow-hidden rounded border border-slate-200 bg-white">
      <table class="w-full text-sm">
        <thead>
          <tr class="bg-slate-100 text-slate-700">
            <th class="px-3 py-2 text-left">ID</th>
            <th class="px-3 py-2 text-left">제목</th>
            <th class="px-3 py-2 text-center">활성</th>
            <th class="px-3 py-2 text-center">순서</th>
          </tr>
        </thead>
        {#if isBusy}
          <thead>
            <tr>
              <th colspan={4} class="p-0!">
                <div class="progress-wrap">
                  <div class="progress-bar"></div>
                </div>
              </th>
            </tr>
          </thead>
        {/if}
        <tbody>
          {#each orderedRows as row (row.id)}
            <tr class="border-t border-slate-200">
              <td class="px-3 py-2 text-slate-700">{row.id}</td>
              <td class="px-3 py-2 text-slate-800">
                <a href={toWikiPath(row.title)} target="_blank" rel="noopener noreferrer external" class="hover:underline">
                  {row.title}
                </a>
              </td>
              <td class="px-3 py-2 text-center text-slate-700">
                <div class="inline-flex items-center justify-center">
                  <ZToggle
                    checked={row.enabled}
                    label={`${row.title} enabled`}
                    disabled={!isSysop}
                    onchange={(event) => void updateEnabled(row.id, event.checked)}
                  />
                </div>
              </td>
              <td class="px-3 py-2 text-center text-slate-700">
                {#if row.enabled}
                  <div class="inline-flex items-center gap-2">
                    <button
                      type="button"
                      aria-label="올리기"
                      class="rounded border border-slate-300 px-2 py-1 text-xs text-slate-700 hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-50"
                      disabled={!isSysop || enabledIndex(row.id) <= 0}
                      onclick={() => void moveEnabled(row.id, -1)}
                    >
                      <ZIcon path={mdiArrowUp} size={14} />
                    </button>
                    <button
                      type="button"
                      aria-label="내리기"
                      class="rounded border border-slate-300 px-2 py-1 text-xs text-slate-700 hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-50"
                      disabled={!isSysop || enabledIndex(row.id) < 0 || enabledIndex(row.id) >= enabledOrder.length - 1}
                      onclick={() => void moveEnabled(row.id, 1)}
                    >
                      <ZIcon path={mdiArrowDown} size={14} />
                    </button>
                  </div>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .progress-wrap {
    position: relative;
    overflow: hidden;
    height: 4px;
    background: var(--background-color-progressive-subtle);
  }

  .progress-bar {
    position: absolute;
    inset-block: 0;
    inline-size: 40%;
    animation: article-tpl-progress 1s infinite linear;
    background: var(--color-progressive);
  }

  @keyframes article-tpl-progress {
    from {
      transform: translateX(-100%);
    }
    to {
      transform: translateX(250%);
    }
  }
</style>
