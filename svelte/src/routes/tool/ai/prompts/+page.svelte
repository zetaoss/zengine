<svelte:options runes={true} />

<script lang="ts">
  import { mdiDelete, mdiPlus } from '@mdi/js'

  import { resolve } from '$app/paths'
  import useAuthStore from '$lib/stores/auth'
  import AvatarUser from '$lib/components/avatar/AvatarUser.svelte'
  import CBadge from '$lib/ui/CBadge.svelte'
  import CButton from '$shared/ui/CButton.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import { showToast } from '$shared/ui/toast/toast'
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
    created_at?: string
    updated_at?: string
  }

  interface BuiltinChain {
    name: string
    description: string
    metadata?: Record<string, string>
    inputs: string[]
    steps: Array<{
      name: string
      description: string
      prompt: { inputs: string[]; messages: Array<{ role: string; content: string }> }
      parser: string
      output: string
    }>
  }

  let builtinChains = $state<BuiltinChain[]>([])

  async function fetchBuiltinChains() {
    const [data, err] = await httpy.get<BuiltinChain[]>('/api/ai/builtin-chains')
    if (err) {
      console.error(err)
      return
    }
    builtinChains = data ?? []
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
      void fetchBuiltinChains()
    }
  })

  async function fetchPromptList() {
    promptListLoading = true
    const [data, err] = await httpy.get<PromptItem[]>('/api/ai-prompts')
    promptListLoading = false
    if (err) {
      console.error(err)
      return
    }
    promptList = (data ?? []).sort((a, b) => b.id - a.id)
  }

  function getRequestTypeLabel(requestType: string) {
    if (requestType === 'create') return '생성'
    if (requestType === 'update') return '편집'
    return requestType || '-'
  }

  function getRequestTypeClass(requestType: string) {
    if (requestType === 'create') return 'text-a-emerald-600'
    if (requestType === 'update') return 'text-a-amber-600'
    return ''
  }

  async function delPrompt(item: PromptItem) {
    const ok = await showConfirm(`'${item.title}' 프롬프트를 삭제하시겠습니까?`)
    if (!ok) return

    deletingPromptId = item.id
    const [, err] = await httpy.delete(`/api/ai-prompts/${item.id}`)
    deletingPromptId = null
    if (err) {
      showToast(err.message || '삭제 실패')
      return
    }

    showToast('삭제 완료')
    promptList = promptList.filter((row) => row.id !== item.id)
  }
</script>

<div class="p-5">
  <div class="mb-3 flex items-center justify-between">
    <h2 class="text-base font-semibold">커스텀</h2>
    {#if isSysop}
      <CButton variant="outline" size="small" href={resolve('/tool/ai/prompts/new')}>
        <ZIcon path={mdiPlus} class="mr-1" />
        새 프롬프트 작성
      </CButton>
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
          <td colspan="5" class="text-center text-muted-foreground">등록된 프롬프트가 없습니다.</td>
        </tr>
      {:else}
        {#each promptList as item (item.id)}
          <tr>
            <td class="text-center text-sm text-muted-foreground">{item.id}</td>
            <td>
              <CBadge variant="outline" class={`mr-2 ${getRequestTypeClass(item.request_type)}`}
                >{getRequestTypeLabel(item.request_type)}</CBadge
              >
              <a class="font-medium hover:underline" href={resolve(`/tool/ai/prompts/${item.id}` as '/tool/ai/prompts/[id]')}>{item.title}</a>
            </td>
            <td class="text-center text-sm text-muted-foreground">{item.use_count === 0 ? '-' : item.use_count}</td>
            <td>
              <AvatarUser user={{ id: item.user_id, name: item.user_name }} />
            </td>
            <td class="text-center text-sm text-muted-foreground">
              {item.created_at?.substring(0, 10) || '-'}
            </td>
            <td class="text-center text-sm text-muted-foreground">
              {item.updated_at?.substring(0, 10) || '-'}
            </td>
            <td class="text-center">
              <div class="flex items-center justify-center gap-1">
                {#if isSysop}
                  <CButton variant="ghost" size="small" disabled={deletingPromptId === item.id} onclick={() => void delPrompt(item)}>
                    <ZIcon path={mdiDelete} />
                  </CButton>
                {/if}
              </div>
            </td>
          </tr>
        {/each}
      {/if}
    </tbody>
  </table>

  {#if builtinChains.length > 0}
    <div class="mt-8 space-y-2">
      <h2 class="text-base font-semibold">빌트인</h2>
      {#each builtinChains as c (c.name)}
        <details class="rounded border bg-card">
          <summary class="cursor-pointer px-3 py-2 text-sm">
            <span class="font-mono text-xs">{c.name}</span>
            <span class="ml-1 font-mono text-xs text-muted-foreground"
              >{Object.entries(c.metadata ?? {})
                .map(([k, v]) => `${k}=${v}`)
                .join(' ')}</span
            >
            <span class="ml-1 font-mono text-xs text-muted-foreground">{c.steps.map((st) => st.name).join(' → ')}</span>
            <span class="ml-2 text-xs text-muted-foreground">{c.description}</span>
          </summary>
          <div class="divide-y border-t">
            {#each c.steps as st, n (st.name)}
              <div class="space-y-2 p-3">
                <div class="text-sm">
                  <span class="text-muted-foreground">{n + 1}.</span>
                  <span class="font-mono text-xs">{c.name}/{st.name}</span>
                  <span class="ml-2 text-xs text-muted-foreground">{st.description}</span>
                </div>
                <div class="flex flex-wrap items-center gap-1.5 font-mono text-xs text-muted-foreground">
                  <span class="rounded bg-muted px-1.5 py-0.5 text-foreground">{st.prompt.inputs.join(', ')}</span>
                  <span>→</span>
                  <span class="rounded bg-muted/60 px-1.5 py-0.5">prompt | model | {st.parser} parser</span>
                  <span>→</span>
                  <span class="rounded bg-muted px-1.5 py-0.5 text-foreground">{st.output}</span>
                </div>
                {#each st.prompt.messages as m (m.role)}
                  <div class="rounded border">
                    <div class="border-b bg-muted/50 px-3 py-1 font-mono text-xs text-muted-foreground">{m.role}</div>
                    <pre class="max-h-[60vh] overflow-auto whitespace-pre-wrap bg-muted/30 p-3 text-xs">{m.content}</pre>
                  </div>
                {/each}
              </div>
            {/each}
          </div>
        </details>
      {/each}
    </div>
  {/if}
</div>
