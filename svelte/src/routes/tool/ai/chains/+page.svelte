<svelte:options runes={true} />

<script lang="ts">
  import { onMount } from 'svelte'

  import { mdiChevronDown, mdiContentCopy, mdiDelete, mdiPencil, mdiPlus } from '@mdi/js'

  import { resolve } from '$app/paths'
  import useAuthStore from '$lib/stores/auth'
  import CButton from '$shared/ui/CButton.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import httpy from '$shared/utils/httpy'

  import { fetchModelIds, formatModelLabel } from '../loop'

  interface ChainMessage {
    role: string
    content: string
    if?: string
    each?: string
  }

  interface ChainStep {
    name: string
    description: string
    inputs: string[]
    messages: ChainMessage[]
    format?: string
    output?: string
  }

  interface ChainInput {
    name: string
    label?: string
    type?: string
  }

  interface Chain {
    name: string
    metadata: { version: string; description: string }
    spec: { userInputs?: ChainInput[]; inputs: ChainInput[]; steps: ChainStep[] }
    model: string
    uses: number
  }

  function inputLabel(input: ChainInput) {
    const type = input.type && input.type !== 'string' ? `: ${input.type}` : ''
    return input.label ? `${input.label}(${input.name}${type})` : `${input.name}${type}`
  }

  function messageHeader(m: ChainMessage) {
    return [m.role, m.if && `if ${m.if}`, m.each && `each ${m.each}`].filter(Boolean).join(' · ')
  }

  interface PromptItem {
    id: number
    user_id: number
    user_name: string
    title: string
    request_type: string
    use_count: number
    content: string
    model: string
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
  let cloningPromptId = $state<number | null>(null)
  let savingPromptId = $state<number | null>(null)
  const userInfo = useAuthStore().userInfo

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

  function toggle(key: string) {
    expanded = expanded === key ? null : key
  }

  // The messages a template runs with until templates become chains of their own: the template in place of the
  // step's unconditional messages, then the run form's (if, each). goapp/tasks/aiedit/request.go withTemplate.
  function templateMessages(item: PromptItem, st: ChainStep): ChainMessage[] {
    return [{ role: 'user', content: item.content }, ...st.messages.filter((m) => m.if || m.each)]
  }

  async function setPromptModel(item: PromptItem, model: string) {
    savingPromptId = item.id
    const [, err] = await httpy.post('/api/ai-prompts', {
      id: item.id,
      title: item.title,
      request_type: item.request_type,
      content: item.content,
      model,
    })
    savingPromptId = null
    if (err) {
      showToast(err.message || '모델 저장에 실패했습니다.')
      return
    }
    item.model = model
    showToast('저장했습니다')
  }

  // A clone takes the model its source runs on as a value, so later changes to the source do not follow.
  async function clonePrompt(item: PromptItem, chain: Chain) {
    cloningPromptId = item.id
    const [, err] = await httpy.post('/api/ai-prompts', {
      title: `${item.title} 복사본`,
      request_type: item.request_type,
      content: item.content,
      model: item.model || chain.model,
    })
    cloningPromptId = null
    if (err) {
      showToast(err.message || '복제 실패')
      return
    }
    showToast('복제했습니다')
    await fetchPrompts()
  }

  onMount(() => {
    fetchChains()
    fetchPrompts()
    fetchModelIds().then((ids) => (modelIds = ids))
  })
</script>

{#snippet expandButton(key: string, isExpanded: boolean)}
  <button
    type="button"
    class="text-muted-foreground hover:text-foreground transition p-1 inline-flex items-center justify-center rounded hover:bg-muted"
    aria-label={isExpanded ? '접기' : '펼치기'}
    onclick={(e) => {
      e.stopPropagation()
      toggle(key)
    }}
  >
    <ZIcon path={mdiChevronDown} size={18} class={`transition-transform duration-200 ${isExpanded ? 'rotate-180' : ''}`} />
  </button>
{/snippet}

{#snippet stepView(st: ChainStep, n: number, messages: ChainMessage[])}
  <div class="space-y-1.5 text-xs min-w-0">
    <div>
      <span class="text-muted-foreground">{n + 1}.</span>
      <span class="font-mono font-medium">{st.name}</span>
      <span class="ml-1 font-mono text-muted-foreground">{st.inputs.join(', ')} → {st.format ?? 'text'}{st.output ? ` → ${st.output}` : ''}</span>
      <span class="ml-2 text-muted-foreground">{st.description}</span>
    </div>
    {#each messages as m, i (i)}
      <div class="rounded border">
        <div class="border-b bg-muted/50 px-3 py-1 font-mono text-muted-foreground">{messageHeader(m)}</div>
        <pre class="max-h-80 overflow-auto whitespace-pre-wrap bg-muted/30 p-3 text-[11px]">{m.content}</pre>
      </div>
    {/each}
  </div>
{/snippet}

<div class="container mx-auto p-4 md:p-6 max-w-6xl space-y-6">
  <h1 class="text-2xl font-bold tracking-tight border-b pb-4">체인</h1>

  <div class="rounded-lg border bg-card overflow-hidden shadow-sm">
    <div class="overflow-x-auto">
      <table class="w-full text-left border-collapse text-sm">
        <thead>
          <tr class="border-b bg-muted/50 text-muted-foreground text-xs font-medium">
            <th class="p-3">체인</th>
            <th class="p-3">설명</th>
            <th class="p-3">모델</th>
            <th class="p-3 text-center whitespace-nowrap">사용</th>
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
            <tr class="hover:bg-muted/30 transition cursor-pointer" onclick={() => toggle(chain.name)}>
              <td class="p-3 whitespace-nowrap">
                <span class="text-muted-foreground">System |</span>
                <span class="font-mono font-medium">{chain.name}</span>
                <span class="ml-1 font-mono text-xs text-muted-foreground">{chain.metadata.version}</span>
              </td>
              <td class="p-3 text-muted-foreground">{chain.metadata.description}</td>
              <td class="p-3 text-xs">{formatModelLabel(chain.model)}</td>
              <td class="p-3 text-center text-muted-foreground tabular-nums" title="이 체인이 끝까지 돈 횟수(프롬프트 틀로 돈 것 포함)">{chain.uses || '-'}</td>
              <td class="p-3 text-center">
                {@render expandButton(chain.name, isExpanded)}
              </td>
            </tr>
            {#if isExpanded}
              <tr class="bg-muted/10">
                <td colspan="5" class="p-4">
                  <div class="space-y-4">
                    <label class="flex items-center gap-2 text-xs">
                      <span class="font-medium">모델</span>
                      <select
                        class="h-8 w-full max-w-72 rounded border bg-background px-2.5 text-xs focus:outline-none focus:ring-1 focus:ring-primary"
                        value={chain.model}
                        disabled={savingChain === chain.name}
                        onchange={(e) => setModel(chain, e.currentTarget.value)}
                      >
                        {#each modelOptions(chain) as id (id)}
                          <option value={id}>{formatModelLabel(id)}</option>
                        {/each}
                      </select>
                    </label>
                    <div class="space-y-1 text-xs text-muted-foreground">
                      {#if chain.spec.userInputs?.length}
                        <div>실행 입력: <span class="font-mono">{chain.spec.userInputs.map(inputLabel).join(', ')}</span></div>
                      {/if}
                      <div>task 입력: <span class="font-mono">{chain.spec.inputs.map(inputLabel).join(', ')}</span></div>
                    </div>
                    {#each chain.spec.steps as st, n (st.name)}
                      {@render stepView(st, n, st.messages)}
                    {/each}
                    {#if promptType}
                      <div class="flex justify-end">
                        <CButton variant="outline" size="small" href={resolve('/tool/ai/prompts/new')}>
                          <ZIcon path={mdiPlus} class="mr-1" />
                          새 프롬프트 틀
                        </CButton>
                      </div>
                    {/if}
                  </div>
                </td>
              </tr>
            {/if}
            {#if promptType}
              {#each prompts.filter((p) => p.request_type === promptType) as item (item.id)}
                {@const key = `prompt:${item.id}`}
                {@const itemExpanded = expanded === key}
                <tr class="hover:bg-muted/30 transition cursor-pointer" onclick={() => toggle(key)}>
                  <td class="p-3 whitespace-nowrap">
                    <span class="text-muted-foreground">{item.user_name} |</span>
                    <a
                      class="font-medium hover:underline"
                      href={resolve(`/tool/ai/prompts/${item.id}` as '/tool/ai/prompts/[id]')}
                      onclick={(e) => e.stopPropagation()}>{item.title}</a
                    >
                  </td>
                  <td class="p-3 text-muted-foreground">
                    {chain.name}의 프롬프트 틀
                  </td>
                  <td class="p-3 text-xs">{formatModelLabel(item.model || chain.model)}</td>
                  <td class="p-3 text-center text-muted-foreground tabular-nums">{item.use_count || '-'}</td>
                  <td class="p-3 text-center">
                    {@render expandButton(key, itemExpanded)}
                  </td>
                </tr>
                {#if itemExpanded}
                  <tr class="bg-muted/10">
                    <td colspan="5" class="p-4">
                      <div class="space-y-4">
                        <label class="flex items-center gap-2 text-xs">
                          <span class="font-medium">모델</span>
                          <select
                            class="h-8 w-full max-w-72 rounded border bg-background px-2.5 text-xs focus:outline-none focus:ring-1 focus:ring-primary"
                            value={item.model || chain.model}
                            disabled={item.user_id !== $userInfo?.id || savingPromptId === item.id}
                            title={item.user_id === $userInfo?.id ? '' : '만든 사람만 바꿀 수 있습니다'}
                            onchange={(e) => setPromptModel(item, e.currentTarget.value)}
                          >
                            {#each modelOptions(chain) as id (id)}
                              <option value={id}>{formatModelLabel(id)}</option>
                            {/each}
                          </select>
                        </label>
                        <p class="text-xs text-muted-foreground">
                          틀이 {chain.name}의 조건 없는 메시지(system과 제목·기존 문서)를 대신하고, 실행 입력 메시지(if·each)는 그대로 붙습니다. 틀의
                          변수: {'{title}'}, {'{wikitext}'}(편집).
                        </p>
                        {#each chain.spec.steps as st, n (st.name)}
                          {@render stepView(st, n, templateMessages(item, st))}
                        {/each}
                        <div class="flex justify-end gap-2">
                          {#if item.user_id === $userInfo?.id}
                            <CButton variant="outline" size="small" href={resolve(`/tool/ai/prompts/${item.id}` as '/tool/ai/prompts/[id]')}>
                              <ZIcon path={mdiPencil} class="mr-1" />
                              편집
                            </CButton>
                          {/if}
                          <CButton variant="outline" size="small" disabled={cloningPromptId === item.id} onclick={() => void clonePrompt(item, chain)}>
                            <ZIcon path={mdiContentCopy} class="mr-1" />
                            복제
                          </CButton>
                          <CButton variant="ghost" size="small" disabled={deletingPromptId === item.id} onclick={() => void delPrompt(item)}>
                            <ZIcon path={mdiDelete} />
                          </CButton>
                        </div>
                      </div>
                    </td>
                  </tr>
                {/if}
              {/each}
            {/if}
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>
