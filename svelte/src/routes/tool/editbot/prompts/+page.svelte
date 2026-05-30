<svelte:options runes={true} />

<script lang="ts">
  import { mdiContentCopy, mdiDelete, mdiPlus, mdiStar, mdiStarOutline } from '@mdi/js'

  import { resolve } from '$app/paths'
  import useAuthStore from '$lib/stores/auth'
  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import { showToast } from '$shared/ui/toast/toast'
  import ZBadge from '$shared/ui/ZBadge.svelte'
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZButtonLink from '$shared/ui/ZButtonLink.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'

  interface PromptItem {
    id: number
    user_id: number
    user_name: string
    title: string
    request_type: string
    content: string
    use_count: number
    is_favorite: boolean
    created_at?: string
    updated_at?: string
  }

  let promptList = $state<PromptItem[]>([])
  let promptListLoading = $state(false)
  let deletingPromptId = $state<number | null>(null)
  let promptListFetched = false

  const auth = useAuthStore()
  const userInfo = auth.userInfo
  let isSysop = $derived(($userInfo?.groups ?? []).includes('sysop'))

  $effect(() => {
    if (!promptListFetched) {
      promptListFetched = true
      void fetchPromptList()
    }
  })

  async function fetchPromptList() {
    promptListLoading = true
    const [data, err] = await httpy.get<PromptItem[]>('/api/editbot/prompts')
    promptListLoading = false
    if (err) {
      console.error(err)
      return
    }
    promptList = (data ?? []).sort((a, b) => b.id - a.id)
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

  async function delPrompt(item: PromptItem) {
    const ok = await showConfirm(`'${item.title}' 프롬프트를 삭제하시겠습니까?`)
    if (!ok) return

    deletingPromptId = item.id
    const [, err] = await httpy.delete(`/api/editbot/prompts/${item.id}`)
    deletingPromptId = null
    if (err) {
      showToast(err.message || '삭제 실패')
      return
    }

    showToast('삭제 완료')
    promptList = promptList.filter((row) => row.id !== item.id)
  }

  async function toggleFavorite(item: PromptItem) {
    if (!$userInfo) {
      showToast('로그인이 필요합니다.')
      return
    }

    const [data, err] = await httpy.post<{ is_favorite: boolean }>(`/api/editbot/prompts/${item.id}/favorite`)
    if (err) {
      showToast(err.message || '즐겨찾기 실패')
      return
    }

    item.is_favorite = data?.is_favorite ?? false
    showToast(item.is_favorite ? '즐겨찾기 지정' : '즐겨찾기 해제')
  }

  async function clonePrompt(item: PromptItem) {
    const ok = await showConfirm(`'${item.title}' 프롬프트를 복제하시겠습니까?`)
    if (!ok) return

    const [detail, getErr] = await httpy.get<PromptItem>(`/api/editbot/prompts/${item.id}`)
    if (getErr || !detail) {
      showToast(getErr?.message || '프롬프트 정보를 가져오지 못했습니다.')
      return
    }

    const [, postErr] = await httpy.post<PromptItem>('/api/editbot/prompts', {
      title: `${detail.title} (복제)`,
      request_type: detail.request_type,
      content: detail.content,
    })

    if (postErr) {
      showToast(postErr.message || '복제 실패')
      return
    }

    showToast('프롬프트가 복제되었습니다.')
    void fetchPromptList()
  }
</script>

<div class="p-5">
  <div class="mb-5 flex items-center justify-between">
    <div class="flex-1"></div>
    {#if isSysop}
      <ZButtonLink color="primary" size="small" href={resolve('/tool/editbot/prompts/new')}>
        <ZIcon path={mdiPlus} class="mr-1" />
        새 프롬프트 작성
      </ZButtonLink>
    {/if}
  </div>

  <table class="z-table">
    <thead>
      <tr>
        <th>ID</th>
        <th>제목</th>
        <th class="text-center">사용</th>
        <th>작성자</th>
        <th class="text-center">생성일</th>
        <th class="text-center">수정일</th>
        <th>&nbsp;</th>
      </tr>
    </thead>
    <tbody>
      {#if promptListLoading}
        <tr>
          <td colspan="5">
            <div class="flex h-32 items-center justify-center">
              <ZSpinner />
            </div>
          </td>
        </tr>
      {:else if promptList.length === 0}
        <tr>
          <td colspan="5" class="text-center text-(--color-subtle)">등록된 프롬프트가 없습니다.</td>
        </tr>
      {:else}
        {#each promptList as item (item.id)}
          <tr>
            <td class="text-center text-sm text-(--color-subtle)">{item.id}</td>
            <td>
              <ZBadge text={getRequestTypeLabel(item.request_type)} class={`mr-2 ${getRequestTypeClass(item.request_type)}`} />
              <a class="font-medium hover:underline" href={resolve(`/tool/editbot/prompts/${item.id}` as '/tool/editbot/prompts/[id]')}>{item.title}</a>
            </td>
            <td class="text-center text-sm text-(--color-subtle)">{item.use_count === 0 ? '-' : item.use_count}</td>
            <td>
              <AvatarUser user={{ id: item.user_id, name: item.user_name }} />
            </td>
            <td class="text-center text-sm text-(--color-subtle)">
              {item.created_at?.substring(0, 10) || '-'}
            </td>
            <td class="text-center text-sm text-(--color-subtle)">
              {item.updated_at?.substring(0, 10) || '-'}
            </td>
            <td class="text-center">
              <div class="flex items-center justify-center gap-1">
                <ZButton color="ghost" size="small" title="즐겨찾기" onclick={() => void toggleFavorite(item)}>
                  <ZIcon path={item.is_favorite ? mdiStar : mdiStarOutline} class={item.is_favorite ? 'text-amber-500' : ''} />
                </ZButton>
                <ZButton color="ghost" size="small" title="복제" onclick={() => void clonePrompt(item)}>
                  <ZIcon path={mdiContentCopy} />
                </ZButton>
                {#if isSysop}
                  <ZButton color="ghost" size="small" disabled={deletingPromptId === item.id} onclick={() => void delPrompt(item)}>
                    <ZIcon path={mdiDelete} />
                  </ZButton>
                {/if}
              </div>
            </td>
          </tr>
        {/each}
      {/if}
    </tbody>
  </table>
</div>
