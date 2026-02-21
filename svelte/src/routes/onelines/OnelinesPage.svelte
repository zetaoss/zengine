<svelte:options runes={true} />

<script lang="ts">
  import { mdiDelete } from '@mdi/js'

  import { page } from '$app/state'
  import ThePagination from '$lib/components/pagination/ThePagination.svelte'
  import type { PaginateData } from '$lib/components/pagination/types'
  import { useOnelineDelete } from '$lib/composables/useOnelineDelete'
  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'
  import linkify from '$shared/utils/linkify'

  interface Row {
    id: number
    user_id: number
    user_name: string
    created: string
    message: string
  }

  let rows = $state<Row[]>([])
  let paginateData = $state<PaginateData | null>(null)
  let isLoading = $state(false)
  let loadError = $state<string | null>(null)

  const { del, canDelete } = useOnelineDelete({
    onSuccess: () => {
      fetchList()
    },
  })

  let pageNumber = $derived.by(() => {
    const p = Number(page.url.searchParams.get('page'))
    return Number.isFinite(p) && p > 0 ? p : 1
  })

  let lastPage = 0

  async function fetchList() {
    isLoading = true
    loadError = null

    const [data, err] = await httpy.get<{
      data: Row[]
      current_page: number
      last_page: number
    }>('/api/onelines', { page: pageNumber })

    if (err) {
      console.error(err)
      rows = []
      paginateData = null
      loadError = '목록을 불러오지 못했습니다.'
      isLoading = false
      return
    }

    rows = await Promise.all(
      data.data.map(async (r) => ({
        ...r,
        message: (await linkify([r.message]))[0] ?? '',
      })),
    )
    paginateData = {
      current_page: data.current_page,
      last_page: data.last_page,
      path: '/onelines',
    }
    isLoading = false
  }

  $effect(() => {
    if (pageNumber !== lastPage) {
      lastPage = pageNumber
      void fetchList()
    }
  })
</script>

<div class="p-5">
  <h2 class="my-5 text-2xl font-bold">한줄잡담</h2>

  <div class="z-card">
    {#if isLoading}
      <div class="flex items-center justify-center py-10 text-gray-500">
        <ZSpinner />
      </div>
    {:else if loadError}
      <div class="py-10 text-center text-red-500">
        {loadError}
      </div>
    {:else if rows.length === 0}
      <div class="py-10 text-center text-gray-500">아직 등록된 한줄잡담이 없습니다.</div>
    {:else}
      {#each rows as r (r.id)}
        <div class="border-b px-3 py-3 last:border-b-0">
          <AvatarUser user={{ id: r.user_id, name: r.user_name }} />
          <!-- eslint-disable-next-line svelte/no-at-html-tags -->
          <span class="ml-1">{@html r.message}</span>
          <span class="z-muted2 ml-1 text-xs">{r.created.substring(0, 10)}</span>
          {#if $canDelete(r.user_id)}
            <ZButton color="ghost" class="z-muted3 py-1 align-middle leading-none" on:click={() => del(r)}>
              <ZIcon path={mdiDelete} />
            </ZButton>
          {/if}
        </div>
      {/each}
    {/if}
  </div>

  {#if !isLoading && !loadError}
    <div class="py-4 text-center">
      {#if paginateData}
        <ThePagination {paginateData} />
      {/if}
    </div>
  {/if}
</div>
