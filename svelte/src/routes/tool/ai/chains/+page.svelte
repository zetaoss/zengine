<svelte:options runes={true} />

<script lang="ts">
  import { onMount } from 'svelte'

  import { mdiChevronDown, mdiDelete, mdiPlus } from '@mdi/js'

  import { resolve } from '$app/paths'
  import AvatarUser from '$lib/components/avatar/AvatarUser.svelte'
  import CButton from '$shared/ui/CButton.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import httpy from '$shared/utils/httpy'

  import { fetchModelIds, formatModelLabel } from '../loop'

  interface ChainStep {
    name: string
    description: string
    inputs: string[]
    messages: Array<{ role: string; content: string }>
    parser: string
    output: string
  }

  interface Chain {
    name: string
    metadata: { version: string; description: string }
    spec: { inputs: string[]; steps: ChainStep[] }
    model: string
  }

  interface PromptItem {
    id: number
    user_id: number
    user_name: string
    title: string
    request_type: string
    use_count: number
    updated_at?: string
  }

  // The AI edit request type whose prompt templates a chain runs.
  const promptTypeOf: Record<string, string> = { 'doc-create': 'create', 'doc-update': 'update' }

  let chains = $state<Chain[]>([])
  let modelIds = $state<string[]>([])
  let prompts = $state<PromptItem[]>([])
  let loading = $state(true)
  let savingChain = $state('')
  let expanded = $state<string | null>(null)
  let deletingPromptId = $state<number | null>(null)

  async function fetchChains() {
    const [data, err] = await httpy.get<Chain[]>('/api/ai/chains')
    loading = false
    if (err) {
      showToast(err.message || '체인을 불러오지 못했습니다.')
      return
    }
    chains = data ?? []
  }

  async function fetchPrompts() {
    const [data, err] = await httpy.get<PromptItem[]>('/api/ai-prompts')
    if (err) return
    prompts = (data ?? []).sort((a, b) => b.use_count - a.use_count || b.id - a.id)
  }

  async function setModel(chain: Chain, model: string) {
    savingChain = chain.name
    const [data, err] = await httpy.put<{ model: string }>(`/api/ai/chains/${chain.name}/model`, { model })
    savingChain = ''
    if (err || !data) {
      showToast(err?.message || '모델 저장에 실패했습니다.')
      return
    }
    chain.model = data.model
    showToast('저장했습니다')
  }

  async function delPrompt(item: PromptItem) {
    if (!(await showConfirm(`'${item.title}' 프롬프트를 삭제하시겠습니까?`))) return
    deletingPromptId = item.id
    const [, err] = await httpy.delete(`/api/ai-prompts/${item.id}`)
    deletingPromptId = null
    if (err) {
      showToast(err.message || '삭제 실패')
      return
    }
    showToast('삭제 완료')
    prompts = prompts.filter((p) => p.id !== item.id)
  }

  function modelOptions(chain: Chain): string[] {
    return modelIds.includes(chain.model) ? modelIds : [chain.model, ...modelIds]
  }

  function toggle(chain: Chain) {
    expanded = expanded === chain.name ? null : chain.name
  }

  onMount(() => {
    fetchChains()
    fetchPrompts()
    fetchModelIds().then((ids) => (modelIds = ids))
  })
</script>

<div class="container mx-auto p-4 md:p-6 max-w-6xl space-y-6">
  <h1 class="text-2xl font-bold tracking-tight border-b pb-4">체인</h1>

  <div class="rounded-lg border bg-card overflow-hidden shadow-sm">
    <div class="overflow-x-auto">
      <table class="w-full text-left border-collapse text-sm">
        <thead>
          <tr class="border-b bg-muted/50 text-muted-foreground text-xs font-medium">
            <th class="p-3">체인</th>
            <th class="p-3">설명</th>
            <th class="p-3 text-center whitespace-nowrap">Steps</th>
            <th class="p-3">모델</th>
            <th class="p-3 w-14 text-center">상세</th>
          </tr>
        </thead>
        <tbody class="divide-y">
          {#if loading}
            <tr><td colspan="5" class="p-8 text-center text-muted-foreground">불러오는 중...</td></tr>
          {/if}
          {#each chains as chain (chain.name)}
            {@const isExpanded = expanded === chain.name}
            {@const promptType = promptTypeOf[chain.name]}
            <tr class="hover:bg-muted/30 transition cursor-pointer" onclick={() => toggle(chain)}>
              <td class="p-3 whitespace-nowrap">
                <span class="font-mono font-medium">{chain.name}</span>
                <span class="ml-1 font-mono text-xs text-muted-foreground">{chain.metadata.version}</span>
              </td>
              <td class="p-3 text-muted-foreground">{chain.metadata.description}</td>
              <td class="p-3 text-center text-muted-foreground">{chain.spec.steps.length}</td>
              <td class="p-3" onclick={(e) => e.stopPropagation()}>
                <select
                  class="h-8 w-full max-w-72 rounded border bg-background px-2.5 text-xs focus:outline-none focus:ring-1 focus:ring-primary"
                  value={chain.model}
                  disabled={savingChain === chain.name}
                  aria-label={`${chain.name} 모델`}
                  onchange={(e) => setModel(chain, e.currentTarget.value)}
                >
                  {#each modelOptions(chain) as id (id)}
                    <option value={id}>{formatModelLabel(id)}</option>
                  {/each}
                </select>
              </td>
              <td class="p-3 text-center">
                <button
                  type="button"
                  class="text-muted-foreground hover:text-foreground transition p-1 inline-flex items-center justify-center rounded hover:bg-muted"
                  aria-label={isExpanded ? '접기' : '펼치기'}
                  onclick={(e) => {
                    e.stopPropagation()
                    toggle(chain)
                  }}
                >
                  <ZIcon path={mdiChevronDown} size={18} class={`transition-transform duration-200 ${isExpanded ? 'rotate-180' : ''}`} />
                </button>
              </td>
            </tr>
            {#if isExpanded}
              <tr class="bg-muted/10">
                <td colspan="5" class="p-4">
                  <div class="space-y-4">
                    <div class="space-y-3 min-w-0">
                      {#each chain.spec.steps as st, n (st.name)}
                        <div class="space-y-1.5 text-xs">
                          <div>
                            <span class="text-muted-foreground">{n + 1}.</span>
                            <span class="font-mono font-medium">{st.name}</span>
                            <span class="ml-1 font-mono text-muted-foreground"
                              >{st.inputs.join(', ')} → {st.parser} → {st.output}</span
                            >
                            <span class="ml-2 text-muted-foreground">{st.description}</span>
                          </div>
                          {#if !promptType}
                            {#each st.messages as m (m.role)}
                              <div class="rounded border">
                                <div class="border-b bg-muted/50 px-3 py-1 font-mono text-muted-foreground">{m.role}</div>
                                <pre class="max-h-80 overflow-auto whitespace-pre-wrap bg-muted/30 p-3 text-[11px]">{m.content}</pre>
                              </div>
                            {/each}
                          {/if}
                        </div>
                      {/each}
                    </div>

                    {#if promptType}
                      {@const items = prompts.filter((p) => p.request_type === promptType)}
                      <div class="space-y-2 min-w-0 text-xs">
                        <p class="text-muted-foreground">
                          아래 프롬프트 틀 중 하나를 제목과 기존 문서로 채운 내용이 그대로 모델에 들어갑니다.
                        </p>
                        <div class="flex items-center justify-between">
                          <span class="font-medium">프롬프트 틀 <span class="text-muted-foreground">{items.length}</span></span>
                          <CButton variant="outline" size="small" href={resolve('/tool/ai/prompts/new')}>
                            <ZIcon path={mdiPlus} class="mr-1" />
                            새 프롬프트
                          </CButton>
                        </div>
                        {#if items.length === 0}
                          <div class="text-muted-foreground">등록된 프롬프트가 없습니다.</div>
                        {:else}
                          <table class="w-full text-left">
                            <tbody class="divide-y">
                              {#each items as item (item.id)}
                                <tr>
                                  <td class="py-1.5">
                                    <a
                                      class="font-medium hover:underline"
                                      href={resolve(`/tool/ai/prompts/${item.id}` as '/tool/ai/prompts/[id]')}>{item.title}</a
                                    >
                                  </td>
                                  <td class="py-1.5 text-right text-muted-foreground w-12" title="사용 횟수">{item.use_count || '-'}</td>
                                  <td class="py-1.5 pl-3 w-32"><AvatarUser user={{ id: item.user_id, name: item.user_name }} /></td>
                                  <td class="py-1.5 w-10 text-right">
                                    <CButton
                                      variant="ghost"
                                      size="small"
                                      disabled={deletingPromptId === item.id}
                                      onclick={() => void delPrompt(item)}
                                    >
                                      <ZIcon path={mdiDelete} />
                                    </CButton>
                                  </td>
                                </tr>
                              {/each}
                            </tbody>
                          </table>
                        {/if}
                      </div>
                    {/if}
                  </div>
                </td>
              </tr>
            {/if}
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>
