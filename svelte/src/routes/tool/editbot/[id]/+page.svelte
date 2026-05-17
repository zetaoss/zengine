<svelte:options runes={true} />

<script lang="ts">
  import { mdiArrowLeft } from '@mdi/js'

  import { page } from '$app/state'
  import RouteLinkButton from '$lib/components/RouteLinkButton.svelte'
  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import ZBadge from '$shared/ui/ZBadge.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'

  interface DocTask {
    id: number
    user_id: number
    user_name: string
    title: string
    request_type: string
    llm_input: string | null
    llm_output: string | null
    llm_model: string | null
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

  function getDateText(value: string) {
    return value.substring(0, 10)
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
    if (requestType === 'create') return 'text-emerald-600 dark:text-emerald-300'
    if (requestType === 'edit') return 'text-amber-600 dark:text-amber-300'
    return ''
  }

  function getTaskPhaseText(phase: string) {
    return phase
  }
</script>

<div class="p-5">
  <div class="mb-4">
    <RouteLinkButton to="/tool/editbot" color="ghost" size="small">
      <ZIcon path={mdiArrowLeft} />
      작업목록
    </RouteLinkButton>
  </div>

  {#if loading}
    <div class="flex h-32 items-center justify-center">
      <ZSpinner />
    </div>
  {:else if row}
    <h2 class="my-5 text-2xl font-bold">
      <ZBadge text={getRequestTypeLabel(row.request_type)} class={`mr-2 align-middle ${getRequestTypeClass(row.request_type)}`} />
      {row.title}
    </h2>

    <table class="z-table mb-5">
      <tbody>
        <tr>
          <th class="w-32 text-left">번호 / 상태</th>
          <td>#{row.id} <span class="mx-2 text-(--color-subtle)">|</span> {getTaskPhaseText(row.phase)}</td>
        </tr>
        <tr>
          <th class="text-left">등록</th>
          <td class="flex items-center gap-2">
            <AvatarUser user={getUser(row)} />
            <span class="text-(--color-subtle)">{getDateText(row.created_at)}</span>
          </td>
        </tr>
        <tr>
          <th class="text-left">요청 유형</th>
          <td>{row.request_type}</td>
        </tr>
        <tr>
          <th class="text-left">처리 통계</th>
          <td>
            시도 {row.attempts}
            <span class="ml-1 text-(--color-subtle)">(실패 {row.error_count}, skip {row.skip_count})</span>
          </td>
        </tr>
        <tr>
          <th class="text-left">시간 정보</th>
          <td>
            created {formatDateTime(row.created_at)}
            <span class="mx-2 text-(--color-subtle)">|</span>
            updated {formatDateTime(row.updated_at)}
            {#if row.retry_at}
              <span class="mx-2 text-(--color-subtle)">|</span>
              retry {formatDateTime(row.retry_at)} ({formatRemaining(row.retry_at)} 남음)
            {/if}
          </td>
        </tr>
        <tr>
          <th class="text-left">결과 / 모델</th>
          <td>
            {#if row.last_error}
              <span class="text-(--color-subtle)">error</span>
              <span class="mx-1 text-(--color-subtle)">|</span>
              <span>{row.llm_model || '-'}</span>
              <div class="mt-1 whitespace-pre-wrap text-sm text-(--color-subtle)">
                {row.last_error}
              </div>
            {:else}
              <span>ok</span>
              <span class="mx-1 text-(--color-subtle)">|</span>
              <span>{row.llm_model || '-'}</span>
            {/if}
          </td>
        </tr>
      </tbody>
    </table>

    <section class="mb-5">
      <h3 class="mb-2 text-lg font-semibold">입력 프롬프트</h3>
      <textarea
        class="z-input h-[260px] w-full resize-y font-mono text-sm"
        readonly
        value={row.llm_input || ''}
        placeholder="저장된 입력 프롬프트가 없습니다."
      ></textarea>
    </section>

    <section>
      <h3 class="mb-2 text-lg font-semibold">결과물</h3>
      <textarea
        class="z-input h-[400px] w-full resize-y font-mono text-sm"
        readonly
        value={row.llm_output || ''}
        placeholder="내용이 없습니다."
      ></textarea>
    </section>
  {:else}
    <div class="text-(--color-subtle)">편집봇 항목을 찾을 수 없습니다.</div>
  {/if}
</div>

<style>
</style>
