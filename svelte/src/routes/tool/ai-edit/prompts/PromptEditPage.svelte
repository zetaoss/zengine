<svelte:options runes={true} />

<script lang="ts">
  import { mdiArrowLeft, mdiDelete, mdiPencil, mdiStar, mdiStarOutline } from '@mdi/js'

  import { goto } from '$app/navigation'
  import { resolve } from '$app/paths'
  import { page } from '$app/state'
  import RouteLinkButton from '$lib/components/RouteLinkButton.svelte'
  import useAuthStore from '$lib/stores/auth'
  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import CBadge from '$shared/ui/CBadge.svelte'
  import CButton from '$shared/ui/CButton.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSelect from '$shared/ui/ZSelect.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import ZValidator from '$shared/ui/ZValidator.svelte'
  import httpy from '$shared/utils/httpy'

  interface PromptItem {
    id: number
    user_id: number
    user_name: string
    title: string
    request_type: string
    content: string
    is_favorite: boolean
    updated_at?: string
  }

  let row = $state<PromptItem | null>(null)
  let loading = $state(true)
  let isEditing = $state(false)
  let isSaving = $state(false)
  let deleting = $state(false)
  let titleError = $state('')
  let observedId = -1

  const auth = useAuthStore()
  const userInfo = auth.userInfo
  let isSysop = $derived(($userInfo?.groups ?? []).includes('sysop'))

  let id = $derived.by(() => {
    if (page.url.pathname.endsWith('/new')) return 0
    const n = Number(page.params.id)
    return Number.isFinite(n) && n >= 0 ? n : -1
  })

  $effect(() => {
    if (id === observedId) return
    observedId = id
    titleError = ''
    if (id === 0) {
      loading = false
      row = {
        id: 0,
        title: '',
        request_type: 'create',
        content: '',
        is_favorite: false,
        user_id: $userInfo?.id ?? 0,
        user_name: $userInfo?.name ?? '',
      }
      isEditing = true
      return
    }
    row = null
    isEditing = false
    void fetchPrompt()
  })

  $effect(() => {
    const title = row?.title?.trim()
    if (!isEditing || !title) {
      titleError = ''
      return
    }

    const timer = setTimeout(() => {
      void checkTitleExists()
    }, 500)

    return () => clearTimeout(timer)
  })

  function canEdit() {
    if (!row || !$userInfo) return false
    return row.id === 0 || row.user_id === $userInfo.id
  }

  async function fetchPrompt() {
    loading = true
    const [data, err] = await httpy.get<PromptItem>(`/api/ai-edit/prompts/${id}`)
    loading = false
    if (err) {
      console.error(err)
      return
    }
    row = data
    isEditing = false
  }

  async function toggleFavorite() {
    if (!row || !$userInfo) {
      showToast('로그인이 필요합니다.')
      return
    }

    const [data, err] = await httpy.post<{ is_favorite: boolean }>(`/api/ai-edit/prompts/${row.id}/favorite`)
    if (err) {
      showToast(err.message || '즐겨찾기 실패')
      return
    }

    row.is_favorite = data?.is_favorite ?? false
    showToast(row.is_favorite ? '즐겨찾기 지정' : '즐겨찾기 해제')
  }

  function getRequestTypeLabel(requestType: string) {
    if (requestType === 'create') return '생성'
    if (requestType === 'edit') return '편집'
    return requestType || '-'
  }

  function getRequestTypeClass(requestType: string) {
    if (requestType === 'create') return 'text-a-emerald-600'
    if (requestType === 'edit') return 'text-a-amber-600'
    return ''
  }

  function startEdit() {
    if (!canEdit()) return
    isEditing = true
  }

  async function save() {
    if (!row || !row.title.trim()) return
    if (titleError) {
      showToast('입력 내용을 확인해 주세요.')
      return
    }
    isSaving = true

    if (row.id === 0 && $userInfo) {
      row.user_id = $userInfo.id
      row.user_name = $userInfo.name
    }

    const [data, err] = await httpy.post<PromptItem>('/api/ai-edit/prompts', row as unknown as Record<string, unknown>)
    isSaving = false
    if (err) {
      if (err.code === 409) {
        titleError = err.message
      } else {
        showToast(err.message || '저장 실패')
      }
      return
    }
    row = data
    isEditing = false
    showToast('저장 완료')
    await goto(resolve('/tool/ai-edit/prompts'))
  }

  async function delPrompt() {
    if (!row || row.id < 1) return
    const ok = await showConfirm(`'${row.title}' 프롬프트를 삭제하시겠습니까?`)
    if (!ok) return

    deleting = true
    const [, err] = await httpy.delete(`/api/ai-edit/prompts/${row.id}`)
    deleting = false
    if (err) {
      showToast(err.message || '삭제 실패')
      return
    }
    row = null
    showToast('삭제 완료')
    await goto(resolve('/tool/ai-edit/prompts'))
  }

  async function checkTitleExists() {
    if (!row || !row.title.trim()) {
      titleError = ''
      return
    }

    const [data, err] = await httpy.get<{ exists: boolean }>('/api/ai-edit/prompts/exists', {
      title: row.title,
      exclude_id: row.id,
    })

    if (err) return

    if (data?.exists) {
      titleError = '이미 사용 중인 제목입니다.'
    } else {
      titleError = ''
    }
  }
</script>

<div class="p-5">
  <div class="mb-4">
    <RouteLinkButton to="/tool/ai-edit/prompts" variant="ghost" size="small">
      <ZIcon path={mdiArrowLeft} />
      프롬프트 목록
    </RouteLinkButton>
  </div>

  {#if loading}
    <div class="flex h-32 items-center justify-center">
      <ZSpinner />
    </div>
  {:else if row}
    {@const currentRow = row}
    <section class="rounded-lg border border-border bg-muted p-6">
      <div class="mb-4 flex flex-wrap items-center justify-between gap-4">
        <div class="flex flex-1 items-start gap-3">
          {#if isEditing}
            <ZValidator error={titleError} containerClass="flex-1">
              {#snippet left()}
                <ZSelect
                  bind:value={currentRow.request_type}
                  items={[
                    { value: 'create', label: '생성' },
                    { value: 'edit', label: '편집' },
                  ]}
                  class="w-24 shrink-0"
                />
              {/snippet}
              {#snippet children(hasError)}
                <input
                  type="text"
                  bind:value={currentRow.title}
                  class="z-input text-lg font-semibold {hasError ? 'border-a-red-500! focus:ring-a-red-500/20!' : ''}"
                  placeholder="프롬프트 제목"
                />
              {/snippet}
            </ZValidator>
          {:else}
            <div class="flex items-center gap-3 py-1">
              <CBadge class={`text-base ${getRequestTypeClass(row.request_type)}`}>{getRequestTypeLabel(row.request_type)}</CBadge>
              <h2 class="text-2xl font-bold">{row.title}</h2>
            </div>
          {/if}
        </div>

        <div class="flex shrink-0 items-center gap-2">
          {#if isEditing}
            <CButton variant="default" size="small" disabled={isSaving} onclick={() => void save()}>
              {isSaving ? '저장 중...' : '저장'}
            </CButton>
            <CButton
              variant="ghost"
              size="small"
              disabled={isSaving}
              onclick={() => (id === 0 ? void goto(resolve('/tool/ai-edit/prompts')) : void fetchPrompt())}
            >
              취소
            </CButton>
          {:else}
            <CButton variant="ghost" size="small" title="즐겨찾기" onclick={() => void toggleFavorite()}>
              <ZIcon path={row.is_favorite ? mdiStar : mdiStarOutline} class={row.is_favorite ? 'text-a-amber-500' : ''} />
            </CButton>
            {#if canEdit()}
              <CButton variant="outline" size="small" onclick={startEdit}>
                <ZIcon path={mdiPencil} class="mr-1" />
                편집
              </CButton>
            {/if}
            {#if isSysop && row.id > 0}
              <CButton variant="ghost" size="small" disabled={deleting} onclick={() => void delPrompt()}>
                <ZIcon path={mdiDelete} class="mr-1" />
                삭제
              </CButton>
            {/if}
          {/if}
        </div>
      </div>

      {#if row.id > 0}
        <div class="mb-4 flex items-center gap-2 text-sm text-muted-foreground">
          <span>작성자:</span>
          <AvatarUser user={{ id: row.user_id, name: row.user_name }} />
        </div>
      {/if}

      {#if isEditing}
        <textarea
          bind:value={row.content}
          class="z-input min-h-[500px] w-full font-mono text-sm leading-relaxed"
          placeholder={'프롬프트 내용을 입력하세요. {제목}, {기존문서} 등의 변수를 사용할 수 있습니다.'}
        ></textarea>
      {:else}
        <div class="min-h-[300px] rounded border border-border bg-background p-4">
          <pre class="whitespace-pre-wrap font-mono text-sm leading-relaxed">{row.content || '내용이 없습니다.'}</pre>
        </div>
      {/if}
    </section>
  {:else}
    <div class="text-muted-foreground">프롬프트를 찾을 수 없습니다.</div>
  {/if}
</div>
