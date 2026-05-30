<svelte:options runes={true} />

<script lang="ts">
  import { mdiArrowLeft, mdiCompare } from '@mdi/js'

  import { page } from '$app/state'
  import RouteLinkButton from '$lib/components/RouteLinkButton.svelte'
  import useAuthStore from '$lib/stores/auth'
  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import ZBadge from '$shared/ui/ZBadge.svelte'
  import ZButtonLink from '$shared/ui/ZButtonLink.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'
  import { getWikiDiffHref } from '$shared/utils/wikiLink'

  interface DocTask {
    id: number
    user_id: number
    user_name: string
    title: string
    request_type: string
    llm_input: string | null
    llm_output: string | null
    llm_model: string | null
    revid?: number
    phase: string
    attempts: number
    error_count: number
    skip_count: number
    last_error: string | null
    retry_at?: string | null
    created_at: string
    updated_at: string
  }

  let row = $state<DocTask | null>(null)
  let loading = $state(true)
  let observedId = 0

  let id = $derived.by(() => {
    const n = Number(page.params.id)
    return Number.isFinite(n) && n > 0 ? n : 0
  })

  $effect(() => {
    if (id !== observedId) {
      observedId = id
      row = null
      if (id > 0) void fetchData()
    }
  })

  async function fetchData() {
    loading = true
    const [data, err] = await httpy.get<DocTask>(`/api/editbot/${id}`)
    loading = false

    if (err) {
      console.error(err)
      return
    }

    row = data
  }

  function getUser(task: DocTask) {
    return { id: task.user_id, name: task.user_name }
  }

  function formatDateTime(value: string | null | undefined) {
    if (!value) return '-'
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return value
    const y = date.getFullYear()
    const m = String(date.getMonth() + 1).padStart(2, '0')
    const d = String(date.getDate()).padStart(2, '0')
    const hh = String(date.getHours()).padStart(2, '0')
    const mm = String(date.getMinutes()).padStart(2, '0')
    const ss = String(date.getSeconds()).padStart(2, '0')
    return `${y}-${m}-${d} ${hh}:${mm}:${ss}`
  }

  function formatRemaining(value: string | null | undefined) {
    if (!value) return '-'
    const ms = new Date(value).getTime() - Date.now()
    if (Number.isNaN(ms)) return '-'
    const total = Math.max(0, Math.ceil(ms / 1000))
    const h = Math.floor(total / 3600)
    const m = Math.floor((total % 3600) / 60)
    const s = total % 60
    if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
    return `${m}:${String(s).padStart(2, '0')}`
  }

  function getRequestTypeLabel(requestType: string) {
    if (requestType === 'create') return '생성'
    if (requestType === 'edit') return '편집'
    return requestType || '-'
  }

  function getRequestTypeClass(requestType: string) {
    if (requestType === 'create') return 'text-emerald-600'
    if (requestType === 'edit') return 'text-amber-600'
    return ''
  }

  const auth = useAuthStore()
  const userInfo = auth.userInfo
  let isAuthorized = $derived(row ? ($userInfo?.groups ?? []).includes('sysop') || row.user_id === $userInfo?.id : false)
</script>

<div class="p-5">
  <div class="mb-4">
    <RouteLinkButton to="/tool/editbot/tasks" color="ghost" size="small">
      <ZIcon path={mdiArrowLeft} />
      작업목록
    </RouteLinkButton>
  </div>

  {#if loading}
    <div class="flex h-32 items-center justify-center">
      <ZSpinner />
    </div>
  {:else if row}
    {#if isAuthorized}
      <div class="mb-5 flex flex-wrap items-center gap-2">
        <span class="text-xl font-bold">#{row.id}</span>
        <ZBadge text={getRequestTypeLabel(row.request_type)} class={getRequestTypeClass(row.request_type)} />
        <span class="text-xl font-bold">{row.title}</span>
        <ZButtonLink
          color="ghost"
          size="small"
          href={getWikiDiffHref(row.title, row.revid)}
          rel="external noopener noreferrer"
          target="_blank"
          title="차이보기"
        >
          <ZIcon path={mdiCompare} />
        </ZButtonLink>
      </div>

      <div class="mb-5 border-y border-(--border-color-subtle) py-4">
        <div class="columns-1 gap-6 md:columns-2">
          <div class="mb-2 break-inside-avoid grid grid-cols-[7rem_1fr] items-start gap-2">
            <div class="text-sm text-(--color-subtle)">등록자</div>
            <div><AvatarUser user={getUser(row)} /></div>
          </div>
          <div class="mb-2 break-inside-avoid grid grid-cols-[7rem_1fr] items-start gap-2">
            <div class="text-sm text-(--color-subtle)">상태</div>
            <div>{row.phase}</div>
          </div>
          <div class="mb-2 break-inside-avoid grid grid-cols-[7rem_1fr] items-start gap-2">
            <div class="text-sm text-(--color-subtle)">모델</div>
            <div>{row.llm_model || '-'}</div>
          </div>
          <div class="mb-2 break-inside-avoid grid grid-cols-[7rem_1fr] items-start gap-2">
            <div class="text-sm text-(--color-subtle)">처리통계</div>
            <div>
              시도 {row.attempts}
              <span class="ml-1 text-(--color-subtle)">(실패 {row.error_count}, skip {row.skip_count})</span>
            </div>
          </div>
          <div class="mb-2 break-inside-avoid grid grid-cols-[7rem_1fr] items-start gap-2">
            <div class="text-sm text-(--color-subtle)">created</div>
            <div>{formatDateTime(row.created_at)}</div>
          </div>
          <div class="mb-2 break-inside-avoid grid grid-cols-[7rem_1fr] items-start gap-2">
            <div class="text-sm text-(--color-subtle)">updated</div>
            <div>
              {formatDateTime(row.updated_at)}
              {#if row.retry_at}
                <span class="ml-2 text-(--color-subtle)">retry {formatDateTime(row.retry_at)} ({formatRemaining(row.retry_at)} 남음)</span>
              {/if}
            </div>
          </div>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-5 md:grid-cols-2">
        <section>
          <h3 class="mb-2 text-lg font-semibold">입력:</h3>
          <textarea
            class="z-input h-[260px] w-full resize-y font-mono text-sm md:h-[600px]"
            readonly
            value={row.llm_input || ''}
            placeholder="저장된 입력 프롬프트가 없습니다."
          ></textarea>
        </section>

        <section>
          <h3 class="mb-2 text-lg font-semibold">결과:</h3>
          <textarea
            class="z-input h-[400px] w-full resize-y font-mono text-sm md:h-[600px]"
            readonly
            value={row.llm_output || ''}
            placeholder="내용이 없습니다."
          ></textarea>
        </section>
      </div>
    {:else}
      <div class="text-(--color-subtle)">권한이 없습니다. 등록자 또는 Sysop만 볼 수 있습니다.</div>
    {/if}
  {:else}
    <div class="text-(--color-subtle)">편집봇 항목을 찾을 수 없습니다.</div>
  {/if}
</div>

<style>
</style>
