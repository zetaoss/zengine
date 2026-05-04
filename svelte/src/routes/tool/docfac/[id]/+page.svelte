<svelte:options runes={true} />

<script lang="ts">
  import { mdiArrowLeft, mdiAutoFix, mdiContentCopy, mdiCreation } from '@mdi/js'
  import { get } from 'svelte/store'

  import { goto } from '$app/navigation'
  import { resolve } from '$app/paths'
  import { page } from '$app/state'
  import RouteLinkButton from '$lib/components/RouteLinkButton.svelte'
  import useAuthStore from '$lib/stores/auth'
  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import { showToast } from '$shared/ui/toast/toast'
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'
  import { getWikiEditHref } from '$shared/utils/wikiLink'

  interface DocTask {
    id: number
    user_id: number
    user_name: string
    title: string
    request_type: string
    content: string | null
    status: string
    attempts: number
    error_count: number
    skip_count: number
    last_error: string | null
    created_at: string
    updated_at: string
  }

  let row = $state<DocTask | null>(null)
  let loading = $state(true)
  let observedId = 0
  const auth = useAuthStore()
  const canWrite = auth.canWrite

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
    const [data, err] = await httpy.get<DocTask>(`/api/doctasks/${id}`)
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

  function getRequestTypeLabel(requestType: string) {
    if (requestType === 'create') return '생성'
    if (requestType === 'edit') return '개선'
    return requestType || '-'
  }

  function getRequestTypeIcon(requestType: string) {
    if (requestType === 'create') return mdiCreation
    if (requestType === 'edit') return mdiAutoFix
    return mdiCreation
  }

  function getActionButtonText(requestType: string) {
    if (requestType === 'edit') return '문서 개선'
    return '문서 생성'
  }

  function loginPath() {
    const searchParams = new URLSearchParams({
      returnto: `:${page.url.pathname}${page.url.search}`,
    })

    return `/login?${searchParams}`
  }

  async function requireAuthConfirm(actionText: string) {
    if (get(canWrite)) return true

    if (
      await showConfirm(`'${actionText}' 기능은 로그인 사용자만 사용할 수 있습니다. 로그인하시겠습니까?`, {
        okText: '로그인',
      })
    ) {
      await goto(resolve(loginPath() as '/login'))
    }

    return false
  }

  async function openWikiEditWithContent(task: DocTask) {
    if (!(await requireAuthConfirm(getActionButtonText(task.request_type)))) return

    const win = window.open(getWikiEditHref(task.title), '_blank')
    if (!win) return

    const content = task.content || ''
    let tries = 0
    const maxTries = 120
    const timer = window.setInterval(() => {
      tries += 1

      try {
        const textarea = win.document.getElementById('wpTextbox1') as HTMLTextAreaElement | null
        if (textarea) {
          textarea.value = content
          textarea.dispatchEvent(new Event('input', { bubbles: true }))
          window.clearInterval(timer)
          return
        }
      } catch (err) {
        console.error(err)
        window.clearInterval(timer)
        return
      }

      if (tries >= maxTries || win.closed) {
        window.clearInterval(timer)
      }
    }, 100)
  }

  async function copyContentToClipboard(task: DocTask) {
    try {
      await navigator.clipboard.writeText(task.content || '')
      showToast('클립보드에 복사했습니다.')
    } catch (err) {
      console.error(err)
      showToast('클립보드 복사에 실패했습니다.')
    }
  }
</script>

<div class="p-5">
  <div class="mb-4">
    <RouteLinkButton to="/tool/docfac" color="ghost" size="small">
      <ZIcon path={mdiArrowLeft} />
      목록
    </RouteLinkButton>
  </div>

  {#if loading}
    <div class="flex h-32 items-center justify-center">
      <ZSpinner />
    </div>
  {:else if row}
    <h2 class="my-5 text-2xl font-bold">{row.title}</h2>

    <table class="mytable z-card mb-5 w-full">
      <tbody>
        <tr>
          <th class="z-base3 w-32 text-left">번호 / 상태</th>
          <td>#{row.id} <span class="mx-2 text-(--color-subtle)">|</span> {row.status}</td>
        </tr>
        <tr>
          <th class="z-base3 text-left">유형</th>
          <td>
            <div class="flex items-center gap-1">
              <ZIcon path={getRequestTypeIcon(row.request_type)} />
              <span>{getRequestTypeLabel(row.request_type)}</span>
            </div>
          </td>
        </tr>
        <tr>
          <th class="z-base3 text-left">등록</th>
          <td class="flex items-center gap-2">
            <AvatarUser user={getUser(row)} />
            <span class="text-(--color-subtle)">{getDateText(row.created_at)}</span>
          </td>
        </tr>
        <tr>
          <th class="z-base3 text-left">처리 통계</th>
          <td>
            시도 {row.attempts}
            <span class="ml-1 text-(--color-subtle)">(실패 {row.error_count}, skip {row.skip_count})</span>
          </td>
        </tr>
        <tr>
          <th class="z-base3 align-top text-left">마지막 결과</th>
          <td>
            {#if row.last_error}
              <div class="mt-1 whitespace-pre-wrap text-sm text-(--color-subtle)">
                {row.last_error}
              </div>
            {:else}
              <div>-</div>
            {/if}
          </td>
        </tr>
      </tbody>
    </table>

    <section>
      <textarea
        class="z-input h-[400px] w-full resize-y font-mono text-sm"
        readonly
        value={row.content || ''}
        placeholder="내용이 없습니다."
      ></textarea>
      <div class="mt-3 flex justify-start gap-2">
        <ZButton color="default" class="gap-1" onclick={() => copyContentToClipboard(row)}>
          <ZIcon path={mdiContentCopy} />
          <span>클립보드 복사</span>
        </ZButton>
        <ZButton color="primary" class="gap-1" title={`${row.title} 편집`} onclick={() => openWikiEditWithContent(row)}>
          <ZIcon path={getRequestTypeIcon(row.request_type)} />
          <span>{getActionButtonText(row.request_type)}</span>
        </ZButton>
      </div>
    </section>
  {:else}
    <div class="text-(--color-subtle)">문서공장 항목을 찾을 수 없습니다.</div>
  {/if}
</div>

<style>
  th,
  td {
    padding: 0.5rem 1rem;
  }
</style>
