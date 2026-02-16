<svelte:options runes={true} />

<script lang="ts">
  import { mdiCheckBold, mdiContentCopy } from '@mdi/js'
  import { onDestroy } from 'svelte'

  import { page } from '$app/state'
  import RouteLinkButton from '$lib/components/RouteLinkButton.svelte'
  import useAuthStore from '$lib/stores/auth'
  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import { showToast } from '$shared/ui/toast/toast'
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import httpy from '$shared/utils/httpy'

  import TheStar from './TheStar.svelte'
  import type { Row } from './types'
  import { getRatio, getScore, getWikitextTable } from './utils'

  type ActionKey = 'wiki' | 'html' | 'url'

  interface CopyAction {
    key: ActionKey
    label: string
    run: () => Promise<void> | void
  }

  const auth = useAuthStore()
  const canDelete = auth.canDelete

  let row = $state<Row | null>(null)
  let tableEl = $state<HTMLTableElement | null>(null)
  let activeTooltip = $state<ActionKey | null>(null)
  let hideTooltipTimer: ReturnType<typeof setTimeout> | null = null
  let retryTimer: ReturnType<typeof setTimeout> | null = null

  let currentId = $state(0)
  let observedId = 0

  let id = $derived.by(() => {
    const raw = page.params.id
    const n = Number(raw)
    return Number.isFinite(n) && n > 0 ? n : 0
  })

  $effect(() => {
    if (id !== observedId) {
      observedId = id
      currentId = id
      row = null
      clearRetryTimer()
      if (currentId > 0) {
        void fetchDataWithRetry()
      }
    }
  })

  onDestroy(() => {
    clearRetryTimer()
    clearTooltipTimer()
  })

  function clearRetryTimer() {
    if (retryTimer) {
      clearTimeout(retryTimer)
      retryTimer = null
    }
  }

  function clearTooltipTimer() {
    if (hideTooltipTimer) {
      clearTimeout(hideTooltipTimer)
      hideTooltipTimer = null
    }
  }

  function scheduleTooltipHide() {
    clearTooltipTimer()
    hideTooltipTimer = setTimeout(() => {
      activeTooltip = null
      hideTooltipTimer = null
    }, 3000)
  }

  async function copyToClipboard(text: string) {
    try {
      await navigator.clipboard.writeText(text)
      return true
    } catch (err) {
      console.error('Failed to copy text to clipboard:', err)
      return false
    }
  }

  function getURL() {
    return `https://${window.location.hostname}/tool/common-report/${currentId}`
  }

  async function copyURL() {
    const ok = await copyToClipboard(getURL())
    if (!ok) showToast('복사 실패')
  }

  async function copyTableHTML() {
    if (!tableEl) return

    const cleanHTML = tableEl.outerHTML
      .replace(/ (class|rel|target|style)="[^"]*"/g, '')
      .replace(/<table>/g, '<table>\n')
      .replace(/<\/tr>/g, '</tr>\n')
      .replace(/<a [^>]*>(.*?)<\/a>/g, '$1')

    const ok = await copyToClipboard(cleanHTML)
    if (!ok) showToast('복사 실패')
  }

  async function copyTableWikitext() {
    if (!tableEl || !row) return

    const text = getWikitextTable(tableEl, currentId, getURL(), row.created_at)
    const ok = await copyToClipboard(text)
    if (!ok) showToast('복사 실패')
  }

  const copyActions: CopyAction[] = [
    { key: 'wiki', label: '표 복사 (WikiText)', run: copyTableWikitext },
    { key: 'html', label: '표 복사 (HTML)', run: copyTableHTML },
    { key: 'url', label: 'URL 복사', run: copyURL },
  ]

  async function handleCopy(action: CopyAction) {
    await action.run()
    activeTooltip = action.key
    scheduleTooltipHide()
  }

  async function del(r: Row) {
    const label = r.items?.[0]?.name ?? '항목'
    const ok = await showConfirm(`'${label}' 등에 관한 #${r.id}번 통용보고서를 삭제하시겠습니까?`)
    if (!ok) return

    const [, err] = await httpy.delete(`/api/common-report/${r.id}`)
    if (err) {
      console.error(err)
      showToast('삭제 실패')
      return
    }

    showToast('삭제 완료')
    window.location.href = '/tool/common-report'
  }

  async function fetchData() {
    if (currentId <= 0) return

    const [data, err] = await httpy.get<Row>(`/api/common-report/${currentId}`)
    if (err) {
      console.error('Error fetching common report data:', err)
      return
    }

    row = data
  }

  async function fetchDataWithRetry(retryDelay = 1000) {
    await fetchData()

    const phase = row?.phase
    if (phase === 'pending' || phase === 'running') {
      clearRetryTimer()
      retryTimer = setTimeout(() => {
        void fetchDataWithRetry(Math.min(retryDelay * 2, 30000))
      }, retryDelay)
    }
  }

  async function rerun(r: Row) {
    const [, err] = await httpy.post(`/api/common-report/${r.id}/rerun`)
    if (err) {
      console.error(err)
      showToast('재실행 실패')
      return
    }

    await fetchDataWithRetry()
  }

  function formatDateTime(value: string) {
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return value

    const y = date.getFullYear()
    const m = String(date.getMonth() + 1).padStart(2, '0')
    const d = String(date.getDate()).padStart(2, '0')
    const hh = String(date.getHours()).padStart(2, '0')
    const mm = String(date.getMinutes()).padStart(2, '0')
    return `${y}-${m}-${d} ${hh}:${mm}`
  }
</script>

{#if row && row.id}
  <div class="p-5">
    <div class="flex justify-end">
      <RouteLinkButton to="/tool/common-report">목록</RouteLinkButton>
    </div>

    <div class="z-card my-2 rounded border p-5">
      <div class="my-5 flex items-center gap-3 text-2xl font-bold">
        <h2 class="m-0">통용 보고서 #{currentId}</h2>
        <div class="flex items-center gap-1 text-base text-gray-600">
          {#if row.phase === 'pending'}
            <span>⏳</span>
          {:else if row.phase === 'running'}
            <span class="spin">⏳</span>
          {:else if row.phase === 'failed'}
            <span>❌</span>
          {:else if row.phase === 'succeeded'}
            <span>✅</span>
          {/if}
          {#if row.phase !== 'succeeded'}
            <span>{row.phase}</span>
          {/if}
        </div>
      </div>

      <div class="flex items-center gap-3">
        <div class="flex-1">
          <AvatarUser user={{ id: row.user_id, name: row.user_name }} />
          <div>{formatDateTime(row.created_at)}</div>
        </div>
        <div class="flex flex-wrap gap-2">
          {#each copyActions as action (action.key)}
            <ZButton onclick={() => handleCopy(action)}>
              <span class="mr-2">{action.label}</span>
              {#if activeTooltip !== action.key}
                <ZIcon path={mdiContentCopy} />
              {:else}
                <span class="inline-flex items-center gap-1 text-green-600">
                  <ZIcon path={mdiCheckBold} />
                  <span class="text-xs">Copied!</span>
                </span>
              {/if}
            </ZButton>
          {/each}
        </div>
      </div>

      <hr class="my-4" />

      <table bind:this={tableEl} class="detail-table border-collapse">
        <tbody>
          <tr>
            <th colspan="2">표기</th>
            {#each row.items as item (item.id)}
              <td>
                <a class="new" href={`/wiki/${item.name}`} rel="external" data-sveltekit-reload>{item.name}</a>
              </td>
            {/each}
          </tr>
          <tr>
            <th colspan="2">판정</th>
            {#each row.items as item, idx (item.id)}
              <td>
                {#if idx === 0}
                  <TheStar n={getScore(row)} />
                {:else}
                  —
                {/if}
              </td>
            {/each}
          </tr>
          <tr>
            <th colspan="2">비율</th>
            {#each row.items as item, idx (item.id)}
              <td>
                {#if getRatio(row, idx)}
                  {(100 * getRatio(row, idx)).toFixed(1)}%
                {:else}
                  —
                {/if}
              </td>
            {/each}
          </tr>
          <tr>
            <th colspan="2">계</th>
            {#each row.items as item (item.id)}
              <td>{item.total.toLocaleString('en-US')}</td>
            {/each}
          </tr>
          <tr>
            <th>다음</th>
            <th>블로그</th>
            {#each row.items as item (item.id)}
              <td>
                <a target="_blank" rel="noopener noreferrer" class="external" href={`http://search.daum.net/search?w=blog&q=${item.name}`}>
                  {item.daum_blog.toLocaleString('en-US')}
                </a>
              </td>
            {/each}
          </tr>
          <tr>
            <th rowspan="3">네이버</th>
            <th>블로그</th>
            {#each row.items as item (item.id)}
              <td>
                <a
                  target="_blank"
                  rel="noopener noreferrer"
                  class="external"
                  href={`https://search.naver.com/search.naver?where=post&query=${item.name}`}
                >
                  {item.naver_blog.toLocaleString('en-US')}
                </a>
              </td>
            {/each}
          </tr>
          <tr>
            <th>책</th>
            {#each row.items as item (item.id)}
              <td>
                <a
                  target="_blank"
                  rel="noopener noreferrer"
                  class="external"
                  href={`http://book.naver.com/search/search.nhn?query=${item.name}`}
                >
                  {item.naver_book.toLocaleString('en-US')}
                </a>
              </td>
            {/each}
          </tr>
          <tr>
            <th>뉴스</th>
            {#each row.items as item (item.id)}
              <td>
                <a
                  target="_blank"
                  rel="noopener noreferrer"
                  class="external"
                  href={`https://search.naver.com/search.naver?where=news&query=${item.name}`}
                >
                  {item.naver_news.toLocaleString('en-US')}
                </a>
              </td>
            {/each}
          </tr>
          <tr>
            <th colspan="2">구글</th>
            {#each row.items as item (item.id)}
              <td>
                <a
                  target="_blank"
                  rel="noopener noreferrer"
                  class="external"
                  href={`http://www.google.com/search?nfpr=1&q=%22${item.name}%22`}
                >
                  {item.google_search.toLocaleString('en-US')}
                </a>
              </td>
            {/each}
          </tr>
        </tbody>
      </table>
    </div>

    <div class="flex gap-2 py-4">
      {#if $canDelete(row.user_id)}
        <ZButton onclick={() => del(row)}>삭제</ZButton>
      {/if}
      {#if $canDelete(row.user_id) && row.phase === 'failed'}
        <ZButton onclick={() => rerun(row)}>재실행</ZButton>
      {/if}
      <div class="flex-1 text-right">
        <RouteLinkButton to="/tool/common-report">목록</RouteLinkButton>
      </div>
    </div>
  </div>
{/if}

<style>
  .detail-table th,
  .detail-table td {
    border: 1px solid currentColor;
    padding: 0.75rem;
    text-align: right;
  }

  .spin {
    display: inline-block;
    animation: common-report-detail-spin 1s linear infinite;
    transform-origin: center;
  }

  @keyframes common-report-detail-spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>
