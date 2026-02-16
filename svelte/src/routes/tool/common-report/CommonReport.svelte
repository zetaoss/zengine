<svelte:options runes={true} />

<script lang="ts">
  import { onDestroy } from 'svelte'
  import { get } from 'svelte/store'

  import { page } from '$app/state'
  import ThePagination from '$lib/components/pagination/ThePagination.svelte'
  import type { PaginateData } from '$lib/components/pagination/types'
  import RouteLinkButton from '$lib/components/RouteLinkButton.svelte'
  import useAuthStore from '$lib/stores/auth'
  import { titlesExist } from '$lib/utils/mediawiki'
  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'

  import CommonReportNew from './CommonReportNew.svelte'
  import { useRetrier } from './retrier'
  import TheStar from './TheStar.svelte'
  import type { Row } from './types'
  import { getRatio, getScore } from './utils'

  interface RespData {
    current_page: number
    data: Row[]
    last_page: number
  }

  const auth = useAuthStore()
  const canWrite = auth.canWrite

  let reportData = $state<RespData | null>(null)
  let paginateData = $state<PaginateData | null>(null)
  let showModal = $state(false)
  let loading = $state(false)
  let titleExists = $state<Record<string, boolean>>({})

  let currentPage = $state(1)
  let observedRoutePage = 0

  const retrier = useRetrier(fetchData)

  let routePage = $derived.by(() => {
    const p = Number(page.url.searchParams.get('page'))
    return Number.isFinite(p) && p > 0 ? p : 1
  })

  $effect(() => {
    if (routePage !== observedRoutePage) {
      observedRoutePage = routePage
      currentPage = routePage
      retrier.start()
    }
  })

  onDestroy(() => retrier.clear())

  async function fetchData() {
    loading = true

    const [data, err] = await httpy.get<RespData>('/api/common-report', {
      page: String(currentPage),
    })

    if (err) {
      console.error('Error fetching common report data:', err)
      loading = false
      return
    }

    reportData = data
    await syncTitleExists(data.data)
    paginateData = {
      current_page: data.current_page,
      last_page: data.last_page,
      path: '/tool/common-report',
    }

    if (data.data.some((row) => ['pending', 'running'].includes(row.phase))) {
      retrier.schedule()
    } else {
      retrier.clear()
    }

    loading = false
  }

  async function syncTitleExists(rows: Row[]) {
    const titles = rows.flatMap((row) => row.items.map((item) => item.name))
    if (titles.length === 0) {
      titleExists = {}
      return
    }

    titleExists = await titlesExist(titles)
  }

  async function openModal() {
    if (!get(canWrite)) {
      window.location.href = '/login?redirect=/tool/common-report'
      return
    }
    showModal = true
  }

  function closeModal() {
    showModal = false
    retrier.start()
  }

  function titleState(name: string, existsMap: Record<string, boolean>): 'unknown' | '' | 'new' {
    const exists = existsMap[name]
    return exists === undefined ? 'unknown' : exists ? '' : 'new'
  }

  function wikiHref(name: string, existsMap: Record<string, boolean>): string {
    return titleState(name, existsMap) === 'new' ? `/wiki/${name}/edit?redlink=1` : `/wiki/${name}`
  }
</script>

<div class="p-5">
  <h2 class="my-5 text-2xl font-bold">통용 보고서</h2>
  <CommonReportNew show={showModal} onClose={closeModal} />

  <table class="w-full">
    <thead class="z-base3">
      <tr>
        <th class="not-mobile">번호</th>
        <th>이름</th>
        <th>건수</th>
        <th>비율</th>
        <th>판정</th>
      </tr>
    </thead>

    {#if loading}
      <thead>
        <tr>
          <th colspan="9" class="p-0!">
            <div class="progress-wrap">
              <div class="progress-bar"></div>
            </div>
            {#if !reportData}
              <div class="flex h-32 items-center justify-center">
                <ZSpinner />
              </div>
            {/if}
          </th>
        </tr>
      </thead>
    {/if}

    {#if reportData}
      {#each reportData.data as row (row.id)}
        <tbody class="report-group align-top border-b border-[#88888866]">
          {#each row.items as item, idx (item.id)}
            <tr>
              {#if idx === 0}
                <td rowspan={row.items.length} class="text-center text-sm">
                  <RouteLinkButton to={`/tool/common-report/${row.id}`}>
                    <span>#{row.id} 상세보기</span>
                  </RouteLinkButton>
                  <div>{row.created_at.substring(0, 10)}</div>
                  <div>
                    <AvatarUser user={{ id: row.user_id, name: row.user_name }} />
                  </div>
                </td>
              {/if}

              <td class="text-right">
                <a href={wikiHref(item.name, titleExists)} rel="external" class={titleState(item.name, titleExists)} data-sveltekit-reload>
                  {item.name}
                </a>
              </td>
              <td class="text-right">
                {item.total.toLocaleString('en-US')}
              </td>
              <td>
                {#if getRatio(row, idx)}
                  <div class="inline-block bg-[#77889966]" style={`width: ${100 * getRatio(row, idx)}%`}>
                    {(100 * getRatio(row, idx)).toFixed(1)}%
                  </div>
                {/if}
              </td>
              <td>
                {#if idx === 0}
                  {#if row.phase === 'pending'}
                    <span>⏳ Pending</span>
                  {:else if row.phase === 'running'}
                    <span class="inline-flex items-center gap-1"><span class="spin">⏳</span> Running</span>
                  {:else if row.phase === 'failed'}
                    <span>❌ Error</span>
                  {:else}
                    <TheStar n={getScore(row)} />
                  {/if}
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      {/each}
    {/if}
  </table>

  <div class="flex justify-end py-2">
    <ZButton disabled={!$canWrite} onclick={openModal}>등록</ZButton>
  </div>

  {#if paginateData}
    <ThePagination {paginateData} />
  {/if}
</div>

<style>
  table :global(th),
  table :global(td) {
    padding: 0.5rem;
  }

  .unknown {
    color: gray;
  }

  .new {
    color: #d53f8c;
  }

  .progress-wrap {
    height: 4px;
    width: 100%;
    overflow: hidden;
    background: rgba(127, 127, 127, 0.2);
  }

  .progress-bar {
    height: 100%;
    width: 35%;
    animation: common-report-progress 1.2s ease-in-out infinite;
    background: #3b82f6;
  }

  .spin {
    display: inline-block;
    animation: common-report-spin 1s linear infinite;
    transform-origin: center;
  }

  @keyframes common-report-progress {
    0% {
      transform: translateX(-120%);
    }
    100% {
      transform: translateX(380%);
    }
  }

  @keyframes common-report-spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }
</style>
