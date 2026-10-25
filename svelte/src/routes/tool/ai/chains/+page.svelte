<svelte:options runes={true} />

<script lang="ts">
  import { onMount } from 'svelte'

  import { mdiChevronDown, mdiContentCopy, mdiDelete, mdiPencil } from '@mdi/js'

  import useAuthStore from '$lib/stores/auth'
  import CButton from '$shared/ui/CButton.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZModal from '$shared/ui/ZModal.svelte'
  import ZTextarea from '$shared/ui/ZTextarea.svelte'
  import httpy from '$shared/utils/httpy'
  import { getAge } from '$shared/utils/time'

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
    used_at?: string | null
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
    description: string
    version: string
    model: string
    used_at?: string | null
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
  let savingPromptId = $state<number | null>(null)
  const now = Date.now()
  const userInfo = useAuthStore().userInfo

  async function fetchChains() {
    const [data, err] = await httpy.get<Chain[]>('/api/ai/chains')
    loading = false
    if (err) {
      showToast(err.message || '체인을 불러오지 못했습니다.')
      return
    }
    // Most recently used first; chains never run keep their order (that of goapp/prompts) after them.
    const usedAt = (c: Chain) => (c.used_at ? new Date(c.used_at).getTime() : 0)
    chains = (data ?? []).map((c, i) => ({ c, i })).sort((a, b) => usedAt(b.c) - usedAt(a.c) || a.i - b.i).map(({ c }) => c)
  }

  // Templates, most recently used first; those not run since used_at was recorded follow by use count.
  async function fetchPrompts() {
    const [data, err] = await httpy.get<PromptItem[]>('/api/ai-prompts')
    if (err) return
    const usedAt = (p: PromptItem) => (p.used_at ? new Date(p.used_at).getTime() : 0)
    prompts = (data ?? []).sort((a, b) => usedAt(b) - usedAt(a) || b.use_count - a.use_count || b.id - a.id)
  }

  // The built-in chain a template runs on: doc-create or doc-update by its request type.
  function chainOf(requestType: string): Chain | undefined {
    return chains.find((c) => promptTypeOf[c.name] === requestType)
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
  function templateMessages(content: string, st: ChainStep): ChainMessage[] {
    return [{ role: 'user', content }, ...st.messages.filter((m) => m.if || m.each)]
  }

  async function setPromptModel(item: PromptItem, model: string) {
    savingPromptId = item.id
    const [, err] = await httpy.post('/api/ai-prompts', {
      id: item.id,
      title: item.title,
      description: item.description,
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

  interface TemplateDraft {
    id: number
    title: string
    description: string
    request_type: string
    content: string
    model: string
    version: string
    // The content when editing began; changing it bumps the version on save.
    original: string
    bump: 'minor' | 'major'
  }

  // nextVersion is prompts.NextVersion: the minor, or with major the major with the minor back to 0.
  function nextVersion(version: string, major: boolean) {
    const [m, n] = version.replace(/^v/, '').split('.').map((x) => Number(x) || 0)
    return major ? `v${m + 1}.0` : `v${m}.${(n ?? 0) + 1}`
  }

  let draft = $state<TemplateDraft | null>(null)
  let draftError = $state('')
  let savingDraft = $state(false)

  function openEdit(item: PromptItem) {
    draftError = ''
    draft = {
      id: item.id,
      title: item.title,
      description: item.description,
      request_type: item.request_type,
      content: item.content,
      model: item.model,
      version: item.version,
      original: item.content,
      bump: 'minor',
    }
    expanded = `prompt:${item.id}`
  }

  // A clone takes the model its source runs on as a value, so later changes to the source do not follow.
  function openClone(item: PromptItem, chain: Chain) {
    draftError = ''
    draft = {
      id: 0,
      title: `${item.title} (복제본)`,
      description: item.description,
      request_type: item.request_type,
      content: item.content,
      model: item.model || chain.model,
      version: 'v1.0',
      original: item.content,
      bump: 'minor',
    }
  }

  // A template cloned from a built-in chain holds the step's unconditional messages (the ones a template takes the
  // place of), so it starts out asking what the chain asks; its model is the chain's, as a value.
  function openCloneChain(chain: Chain, requestType: string) {
    draftError = ''
    const content = chain.spec.steps[0].messages
      .filter((m) => !m.if && !m.each)
      .map((m) => m.content)
      .join('\n\n')
    draft = {
      id: 0,
      title: `${chain.name} (복제본)`,
      description: chain.metadata.description,
      request_type: requestType,
      content,
      model: chain.model,
      version: 'v1.0',
      original: content,
      bump: 'minor',
    }
  }

  function closeDraft() {
    if (!savingDraft) draft = null
  }

  function setDraftContent(value: string) {
    if (draft) draft.content = value
  }

  async function saveDraft() {
    if (!draft || !draft.title.trim()) return
    savingDraft = true
    const body: Record<string, unknown> = {
      id: draft.id,
      title: draft.title.trim(),
      description: draft.description,
      request_type: draft.request_type,
      content: draft.content,
      bump: draft.bump,
    }
    if (draft.id === 0) body.model = draft.model
    const [, err] = await httpy.post('/api/ai-prompts', body)
    savingDraft = false
    if (err) {
      draftError = err.message || '저장 실패'
      return
    }
    showToast(draft.id === 0 ? '복제했습니다' : '저장했습니다')
    draft = null
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

{#snippet templateNote(chain: Chain)}
  <p class="text-xs text-muted-foreground">변수: {'{title}'}{promptTypeOf[chain.name] === 'update' ? `, {wikitext}` : ''}</p>
{/snippet}

{#snippet templateEditor(chain: Chain, d: TemplateDraft)}
  <div class="space-y-4 text-xs">
    <div class="flex flex-wrap items-center gap-2">
      <label class="flex min-w-0 flex-1 items-center gap-2">
        <span class="font-medium">제목</span>
        <input class="h-8 w-full max-w-72 rounded border bg-background px-2.5" bind:value={d.title} />
      </label>
      {#if d.id === 0}
        <label class="flex items-center gap-2">
          <span class="font-medium">모델</span>
          <select class="h-8 max-w-72 rounded border bg-background px-2.5" bind:value={d.model}>
            {#each modelOptions(chain) as id (id)}
              <option value={id}>{formatModelLabel(id)}</option>
            {/each}
          </select>
        </label>
      {/if}
      {#if d.id > 0 && d.content !== d.original}
        <label class="flex items-center gap-2">
          <span class="font-medium">버전</span>
          <select class="h-8 rounded border bg-background px-2.5" bind:value={d.bump}>
            <option value="minor">마이너 {d.version} → {nextVersion(d.version, false)}</option>
            <option value="major">메이저 {d.version} → {nextVersion(d.version, true)}</option>
          </select>
        </label>
      {/if}
      {#if d.id > 0}
        <CButton variant="default" size="small" disabled={!d.title.trim() || savingDraft} onclick={() => void saveDraft()}>
          {savingDraft ? '저장 중...' : '저장'}
        </CButton>
        <CButton variant="outline" size="small" disabled={savingDraft} onclick={closeDraft}>취소</CButton>
      {/if}
    </div>
    <label class="flex items-center gap-2">
      <span class="font-medium">설명</span>
      <input class="h-8 w-full rounded border bg-background px-2.5" bind:value={d.description} />
    </label>
    {#if d.id > 0 && d.content !== d.original && d.bump === 'major'}
      <div class="text-muted-foreground">메이저 범프: 이 틀로 돈 이전 결과와 다른 것으로 구분합니다(AI 편집은 다시 돌리는 대상이 없습니다).</div>
    {/if}
    {#if draftError}
      <div class="text-destructive">{draftError}</div>
    {/if}
    {@render templateNote(chain)}
    {#each chain.spec.steps as st, n (st.name)}
      <div class="space-y-1.5 min-w-0">
        <div>
          <span class="text-muted-foreground">{n + 1}.</span>
          <span class="font-mono font-medium">{st.name}</span>
        </div>
        <div class="rounded border">
          <div class="border-b bg-muted/50 px-3 py-1 font-mono text-muted-foreground">user · 프롬프트 틀</div>
          <div class="font-mono text-[11px]">
            <ZTextarea id={`ai-template-${d.id}`} modelValue={d.content} placeholder={undefined} maxHeight={480} onUpdateModelValue={setDraftContent} />
          </div>
        </div>
        {#each templateMessages(d.content, st).slice(1) as m, i (i)}
          <div class="rounded border">
            <div class="border-b bg-muted/50 px-3 py-1 font-mono text-muted-foreground">{messageHeader(m)}</div>
            <pre class="max-h-80 overflow-auto whitespace-pre-wrap bg-muted/30 p-3 text-[11px]">{m.content}</pre>
          </div>
        {/each}
      </div>
    {/each}
  </div>
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
            <th class="p-3">소유자</th>
            <th class="p-3">체인</th>
            <th class="p-3">설명</th>
            <th class="p-3">모델</th>
            <th class="p-3 text-center whitespace-nowrap">사용</th>
            <th class="p-3 whitespace-nowrap">최근</th>
            <th class="p-3 w-14 text-center">상세</th>
          </tr>
        </thead>
        <tbody class="divide-y">
          {#if loading}
            <tr><td colspan="7" class="p-8 text-center text-muted-foreground">불러오는 중...</td></tr>
          {/if}
          {#each chains as chain (chain.name)}
            {@const isExpanded = expanded === chain.name}
            {@const promptType = promptTypeOf[chain.name]}
            <tr class="bg-muted/50 hover:bg-muted transition cursor-pointer" onclick={() => toggle(chain.name)}>
              <td class="p-3 text-muted-foreground">-</td>
              <td class="p-3 whitespace-nowrap">{chain.name}</td>
              <td class="p-3 text-muted-foreground">{chain.metadata.description}</td>
              <td class="p-3 text-xs">{formatModelLabel(chain.model)}</td>
              <td class="p-3 text-center text-muted-foreground tabular-nums" title="이 체인이 끝까지 돈 횟수(프롬프트 틀로 돈 것 포함)">{chain.uses || '-'}</td>
              <td class="p-3 whitespace-nowrap text-xs text-muted-foreground">{chain.used_at ? getAge(chain.used_at, now) : '-'}</td>
              <td class="p-3 text-center">
                {@render expandButton(chain.name, isExpanded)}
              </td>
            </tr>
            {#if isExpanded}
              <tr class="bg-muted/10">
                <td colspan="7" class="p-4">
                  <div class="space-y-4">
                    <div class="flex flex-wrap items-center gap-2 text-xs">
                      <label class="flex min-w-0 flex-1 items-center gap-2">
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
                      {#if promptType}
                        <CButton variant="outline" size="small" onclick={() => openCloneChain(chain, promptType)}>
                          <ZIcon path={mdiContentCopy} class="mr-1" />
                          복제
                        </CButton>
                      {/if}
                    </div>
                    <div class="space-y-1 text-xs text-muted-foreground">
                      <div>버전: <span class="font-mono">{chain.metadata.version}</span></div>
                      {#if chain.spec.userInputs?.length}
                        <div>실행 입력: <span class="font-mono">{chain.spec.userInputs.map(inputLabel).join(', ')}</span></div>
                      {/if}
                      <div>task 입력: <span class="font-mono">{chain.spec.inputs.map(inputLabel).join(', ')}</span></div>
                    </div>
                    {#each chain.spec.steps as st, n (st.name)}
                      {@render stepView(st, n, st.messages)}
                    {/each}
                  </div>
                </td>
              </tr>
            {/if}
          {/each}
          {#each prompts as item (item.id)}
            {@const chain = chainOf(item.request_type)}
            {#if chain}
                {@const key = `prompt:${item.id}`}
                {@const itemExpanded = expanded === key}
                {@const editing = draft?.id === item.id}
                {@const own = item.user_id === $userInfo?.id}
                <tr
                  class={[
                    'transition cursor-pointer',
                    own ? 'bg-blue-50 hover:bg-blue-100 dark:bg-blue-950/40 dark:hover:bg-blue-900/40' : 'hover:bg-muted/30',
                  ]}
                  onclick={() => toggle(key)}
                >
                  <td class="p-3 whitespace-nowrap text-muted-foreground">{item.user_name}</td>
                  <td class="p-3 whitespace-nowrap">{item.title}</td>
                  <td class="p-3 text-muted-foreground">{item.description}</td>
                  <td class="p-3 text-xs">{formatModelLabel(item.model || chain.model)}</td>
                  <td class="p-3 text-center text-muted-foreground tabular-nums">{item.use_count || '-'}</td>
                  <td class="p-3 whitespace-nowrap text-xs text-muted-foreground">{item.used_at ? getAge(item.used_at, now) : '-'}</td>
                  <td class="p-3 text-center">
                    {@render expandButton(key, itemExpanded)}
                  </td>
                </tr>
                {#if itemExpanded}
                  <tr class="bg-muted/10">
                    <td colspan="7" class="p-4">
                      {#if editing && draft}
                        {@render templateEditor(chain, draft)}
                      {:else}
                        <div class="space-y-4">
                          <div class="flex flex-wrap items-center gap-2 text-xs">
                            <label class="flex min-w-0 flex-1 items-center gap-2">
                              <span class="font-medium">모델</span>
                              <select
                                class="h-8 w-full max-w-72 rounded border bg-background px-2.5 text-xs focus:outline-none focus:ring-1 focus:ring-primary"
                                value={item.model || chain.model}
                                disabled={!own || savingPromptId === item.id}
                                title={own ? '' : '만든 사람만 바꿀 수 있습니다'}
                                onchange={(e) => setPromptModel(item, e.currentTarget.value)}
                              >
                                {#each modelOptions(chain) as id (id)}
                                  <option value={id}>{formatModelLabel(id)}</option>
                                {/each}
                              </select>
                            </label>
                            {#if own}
                              <CButton variant="outline" size="small" onclick={() => openEdit(item)}>
                                <ZIcon path={mdiPencil} class="mr-1" />
                                편집
                              </CButton>
                            {/if}
                            <CButton variant="outline" size="small" onclick={() => openClone(item, chain)}>
                              <ZIcon path={mdiContentCopy} class="mr-1" />
                              복제
                            </CButton>
                            {#if own}
                              <CButton variant="outline" size="small" disabled={deletingPromptId === item.id} onclick={() => void delPrompt(item)}>
                                <ZIcon path={mdiDelete} class="mr-1" />
                                삭제
                              </CButton>
                            {/if}
                          </div>
                          <div class="text-xs text-muted-foreground">버전: <span class="font-mono">{item.version}</span> · 원본: <span class="font-mono">{chain.name}</span></div>
                          {#each chain.spec.steps as st, n (st.name)}
                            {@render stepView(st, n, templateMessages(item.content, st))}
                          {/each}
                        </div>
                      {/if}
                    </td>
                  </tr>
                {/if}
            {/if}
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>

<ZModal
  show={draft !== null && draft.id === 0}
  title="프롬프트 틀 복제"
  okText={savingDraft ? '저장 중...' : '저장'}
  okVariant="default"
  okDisabled={!draft?.title.trim() || savingDraft}
  closable={!savingDraft}
  backdropClosable={!savingDraft}
  panelClass="max-w-[calc(100vw-2rem)] md:max-w-4xl"
  onOk={saveDraft}
  onCancel={closeDraft}
>
  {#if draft && draft.id === 0 && chainOf(draft.request_type)}
    {@render templateEditor(chainOf(draft.request_type)!, draft)}
  {/if}
</ZModal>
