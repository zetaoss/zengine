<svelte:options runes={true} />

<script lang="ts">
  import { onMount } from 'svelte'

  import { mdiChevronDown, mdiContentCopy, mdiPencil, mdiTrashCanOutline } from '@mdi/js'

  import useAuthStore from '$lib/stores/auth'
  import CButton from '$shared/ui/CButton.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import httpy from '$shared/utils/httpy'
  import { getAge } from '$shared/utils/time'

  import { fetchModelIds, formatModelLabel } from '../loop'

  interface ChainMessage {
    role: string
    content: string
    if?: string
    each?: string
  }

  // A step with uses is a function step (no model, no messages); if runs it only when that variable is given.
  interface ChainStep {
    name: string
    description: string
    uses?: string
    if?: string
    inputs: string[]
    messages: ChainMessage[]
    format?: string
    output?: string
  }

  // A variable (prompts.Input). A user input may be required, have a default, and as a choice list its options.
  interface ChainInput {
    name: string
    placeholder?: string
    type?: string
    required?: boolean
    default?: string | boolean
    options?: string[]
    from?: string
    maxChars?: number
  }

  interface Chain {
    name: string
    metadata: { version: string; description: string }
    spec: { userInputs?: ChainInput[]; inputs: ChainInput[]; steps: ChainStep[] }
    model: string
    uses: number
    used_at?: string | null
    template: string
    user_inputs_template?: string
    messages_template?: string
  }

  // A user input's name as the run form shows it: useBacklinks as "Use backlinks" (as in the skin's AiEditRunner).
  function humanize(name: string) {
    const words = name.replace(/([a-z0-9])([A-Z])/g, '$1 $2').toLowerCase()
    return words.charAt(0).toUpperCase() + words.slice(1)
  }

  function inputLabel(input: ChainInput) {
    const type = input.type && input.type !== 'string' ? `: ${input.type}` : ''
    const from = input.from ? ` ← ${input.from}${input.maxChars ? `, ${input.maxChars}자까지` : ''}` : ''
    return `${input.name}${type}${from}`
  }

  type PreviewValue = string | boolean | string[]

  // Values typed into a user inputs preview, by name, to see the messages they make. Kept on screen only.
  let previewValues = $state<Record<string, PreviewValue>>({})
  // The values the task fills for a run (title, body) and the backlinks step's output, fixed in a preview.
  const previewTitle = 'Hello World'
  const previewBody = '문서 시작했다\n문서 끝났다'
  const previewBacklinks = 'C language\nProgramming language'

  // A variable a template always has, shown beside the user inputs it picks: filled by the task (title, body) or by
  // the backlinks step (only when useBacklinks, given as after, is picked). Its preview value is fixed.
  interface FixedVar {
    name: string
    value: string
    after?: string
  }

  function fixedVars(requestType: string): FixedVar[] {
    return [
      { name: 'title', value: previewTitle },
      ...(requestType === 'update' ? [{ name: 'body', value: previewBody }] : []),
      { name: 'backlinks', value: previewBacklinks, after: 'useBacklinks' },
    ]
  }

  function previewTexts(name: string): string[] {
    const v = previewValues[name]
    return Array.isArray(v) ? v : []
  }

  function setPreview(name: string, v: PreviewValue) {
    previewValues = { ...previewValues, [name]: v }
  }

  // A message's content split at its {name} placeholders, as prompts' placeholders finds them (one right after
  // another brace is wiki markup), to show the variables.
  const placeholderRe = /\{([A-Za-z][A-Za-z0-9_]*)\}/g

  function parts(content: string): Array<{ text?: string; name?: string }> {
    const out: Array<{ text?: string; name?: string }> = []
    let last = 0
    for (const m of content.matchAll(placeholderRe)) {
      const at = m.index ?? 0
      if (at > 0 && content[at - 1] === '{') continue
      out.push({ text: content.slice(last, at) }, { name: m[1] })
      last = at + m[0].length
    }
    out.push({ text: content.slice(last) })
    return out
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
    messages: ChainMessage[]
    user_inputs: ChainInput[]
    spec: string
    user_inputs_yaml?: string
    messages_yaml?: string
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
    if (expanded === `prompt:${item.id}`) {
      expanded = null
      draft = null
    }
  }

  function modelOptions(chain: Chain): string[] {
    return modelIds.includes(chain.model) ? modelIds : [chain.model, ...modelIds]
  }


  // A template as the editor shows it: an open row (key) seen as it is, or being edited. id is the template's, 0
  // for a new one (a clone) and -1 for a built-in chain, which is only seen. legacy is a template saved before it
  // had its own messages, whose content becomes its system instructions on save. stored are the messages it runs
  // with, which a row seen previews; an edited one previews the messages a save makes.
  interface TemplateDraft {
    key: string
    editing: boolean
    id: number
    owner: string
    source: string
    legacy: boolean
    title: string
    description: string
    request_type: string
    user_inputs_yaml: string
    messages: ChainMessage[]
    model: string
    version: string
    stored: ChainMessage[]
    original: string
    bump: 'minor' | 'major'
  }

  // nextVersion is prompts.NextVersion: the minor, or with major the major with the minor back to 0.
  function nextVersion(version: string, major: boolean) {
    const [m, n] = version.replace(/^v/, '').split('.').map((x) => Number(x) || 0)
    return major ? `v${m + 1}.0` : `v${m}.${(n ?? 0) + 1}`
  }

  // The user inputs a template picks, by name, from its user inputs YAML (- hint), as the server writes it.
  function pickedNames(yaml: string): string[] {
    return yaml
      .split('\n')
      .map((line) => line.replace(/^-\s*/, '').trim())
      .filter(Boolean)
  }

  // Picks or drops one of the chain's user inputs; the picked ones keep the chain's order, which the messages follow.
  function setPicked(d: TemplateDraft, chain: Chain, name: string, on: boolean) {
    const names = pickedNames(d.user_inputs_yaml).filter((n) => n !== name)
    if (on) names.push(name)
    d.user_inputs_yaml = (chain.spec.userInputs ?? [])
      .filter((input) => names.includes(input.name))
      .map((input) => `- ${input.name}\n`)
      .join('')
  }

  let draft = $state<TemplateDraft | null>(null)
  let draftError = $state('')
  let savingDraft = $state(false)

  // What a save would change, to tell an edited draft from the template it started as.
  function draftState(d: Pick<TemplateDraft, 'title' | 'description' | 'model' | 'user_inputs_yaml' | 'messages'>) {
    return JSON.stringify([d.title, d.description, d.model, d.user_inputs_yaml, d.messages])
  }

  function draftChanged(d: TemplateDraft) {
    return draftState(d) !== d.original
  }

  function newDraft(d: Omit<TemplateDraft, 'original' | 'bump'>): TemplateDraft {
    return { ...d, original: draftState(d), bump: 'minor' }
  }

  // The system block a template edits: its own, or for a template saved before it had messages, its content.
  function systemOf(msgs: ChainMessage[]): ChainMessage[] {
    const system = msgs.find((m) => m.role === 'system')
    return [{ role: 'system', content: system?.content ?? msgs[0]?.content ?? '' }]
  }

  function promptDraft(item: PromptItem, chain: Chain): TemplateDraft {
    const raw = item.messages?.length ? item.messages : [{ role: 'user', content: item.content || '' }]
    return newDraft({
      key: `prompt:${item.id}`,
      editing: false,
      id: item.id,
      owner: item.user_name,
      source: chain.name,
      legacy: !raw.some((m) => m.role === 'system'),
      title: item.title,
      description: item.description,
      request_type: item.request_type,
      user_inputs_yaml: item.user_inputs_yaml ?? '',
      messages: systemOf(raw),
      stored: raw,
      model: item.model,
      version: item.version,
    })
  }

  function chainDraft(chain: Chain, requestType: string): TemplateDraft {
    return newDraft({
      key: chain.name,
      editing: false,
      id: -1,
      owner: '',
      source: '',
      legacy: false,
      title: chain.name,
      description: chain.metadata.description,
      request_type: requestType,
      user_inputs_yaml: chain.user_inputs_template ?? '',
      messages: systemOf(chain.spec.steps.find((step) => !step.uses)?.messages ?? []),
      stored: chain.spec.steps.find((step) => !step.uses)?.messages ?? [],
      model: chain.model,
      version: chain.metadata.version,
    })
  }

  // The draft an open row shows: a template's, or a built-in chain's that templates clone (doc-create, doc-update).
  function draftFor(key: string): TemplateDraft | null {
    const item = prompts.find((p) => `prompt:${p.id}` === key)
    const chain = item ? chainOf(item.request_type) : chains.find((c) => c.name === key)
    if (item && chain) return promptDraft(item, chain)
    const requestType = chain && promptTypeOf[chain.name]
    return chain && requestType ? chainDraft(chain, requestType) : null
  }

  async function confirmDiscard() {
    return !draft?.editing || !draftChanged(draft) || (await showConfirm('저장하지 않은 변경을 버리시겠습니까?'))
  }

  async function toggle(key: string) {
    if (!(await confirmDiscard())) return
    draftError = ''
    expanded = expanded === key ? null : key
    draft = expanded ? draftFor(expanded) : null
  }

  function startEdit() {
    if (draft) draft.editing = true
  }

  async function cancelEdit() {
    if (savingDraft || !(await confirmDiscard())) return
    draftError = ''
    if (draft?.id === 0) {
      expanded = null
      draft = null
    } else {
      draft = expanded ? draftFor(expanded) : null
    }
  }

  // A clone's title: its source's with the cloning user's name in lower case (doc-update-jmnote).
  function cloneTitle(title: string) {
    const name = ($userInfo?.name ?? '').toLowerCase()
    return name ? `${title}-${name}` : title
  }

  // A clone opens as a new template being edited, at the top of the templates. It takes the model its source runs on
  // as a value, so later changes to the source do not follow.
  async function openClone(source: TemplateDraft) {
    if (!(await confirmDiscard())) return
    draftError = ''
    draft = {
      ...source,
      key: 'new',
      editing: true,
      id: 0,
      owner: $userInfo?.name ?? '',
      source: chainOf(source.request_type)?.name ?? '',
      legacy: false,
      title: cloneTitle(source.title),
      model: source.model || chainOf(source.request_type)?.model || '',
      version: 'v1.0',
      messages: source.messages.map((m) => ({ ...m })),
    }
    expanded = 'new'
  }

  interface SpecCheck {
    ok: boolean
    error?: string
    user_inputs?: ChainInput[]
    messages?: ChainMessage[]
    rendered_messages?: ChainMessage[]
    preview_error?: string
  }

  // The editor's inputs and messages as the server reads them on save (POST /api/ai-prompts/check), shortly after typing stops.
  let specCheck = $state<SpecCheck | null>(null)
  $effect(() => {
    const d = draft
    if (!d) {
      specCheck = null
      return
    }
    const userInputsYaml = d.user_inputs_yaml
    const systemPrompt = d.messages[0]?.content ?? ''
    const stored = d.editing ? null : $state.snapshot(d.stored)
    const requestType = d.request_type
    const values = $state.snapshot(previewValues)
    const title = previewTitle
    const body = previewBody
    const backlinks = previewBacklinks.split('\n')
    let active = true
    const timer = setTimeout(async () => {
      const [data] = await httpy.post<SpecCheck>('/api/ai-prompts/check', {
        request_type: requestType,
        user_inputs_yaml: userInputsYaml,
        ...(stored ? { stored: true, messages: stored } : { auto_messages: true, system_prompt: systemPrompt }),
        preview_values: values,
        preview_title: title,
        preview_body: body,
        preview_backlinks: backlinks,
      })
      if (active && data) {
        specCheck = data
      }
    }, 400)
    return () => {
      active = false
      clearTimeout(timer)
    }
  })

  async function saveDraft() {
    if (!draft || !draft.title.trim()) return
    savingDraft = true
    const body: Record<string, unknown> = {
      id: draft.id,
      title: draft.title.trim(),
      description: draft.description,
      request_type: draft.request_type,
      user_inputs_yaml: draft.user_inputs_yaml,
      auto_messages: true,
      system_prompt: draft.messages[0]?.content ?? '',
      bump: draft.bump,
    }
    body.model = draft.model
    const [data, err] = await httpy.post<PromptItem>('/api/ai-prompts', body)
    savingDraft = false
    if (err || !data) {
      draftError = err?.message || '저장 실패'
      return
    }
    showToast(draft.id === 0 ? '복제했습니다' : '저장했습니다')
    await fetchPrompts()
    expanded = `prompt:${data.id}`
    draft = draftFor(expanded)
  }

  // A rendered cell fills the height its row takes from the editor beside it (md and up) and scrolls within it, so a
  // long preview does not stretch the row; on one column it flows, each message capped.
  const renderedCell = 'space-y-2 md:absolute md:inset-0 md:flex md:flex-col md:gap-2 md:space-y-0 md:overflow-auto'

  // Leaving the page with unsaved edits asks first, as closing a row does.
  $effect(() => {
    if (!draft?.editing || !draftChanged(draft)) return
    const warn = (e: BeforeUnloadEvent) => e.preventDefault()
    window.addEventListener('beforeunload', warn)
    return () => window.removeEventListener('beforeunload', warn)
  })

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

{#snippet templateEditor(chain: Chain, d: TemplateDraft)}
  {@const item = prompts.find((p) => p.id === d.id)}
  {@const own = !!item && item.user_id === $userInfo?.id}
  {@const picked = pickedNames(d.user_inputs_yaml)}
  <div class="space-y-4 text-xs">
    <div class="flex flex-wrap items-start gap-3">
      <div class="min-w-0 flex-1 space-y-2">
        {#if d.editing}
          <input class="h-8 w-full max-w-md rounded border bg-background px-2.5 text-sm font-medium" aria-label="제목" placeholder="제목" bind:value={d.title} />
          <input class="h-8 w-full rounded border bg-background px-2.5" aria-label="설명" placeholder="설명" bind:value={d.description} />
        {:else}
          <div class="flex h-8 items-center text-sm font-medium">{d.title}</div>
          <div class="flex h-8 items-center text-muted-foreground">{d.description || '설명 없음'}</div>
        {/if}
        <div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-muted-foreground">
          {#if d.owner}
            <span>소유자 <span class="text-foreground">{d.owner}</span></span>
          {/if}
          <span class="flex items-center gap-1.5">
            버전 <span class="font-mono text-foreground">{d.version}</span>
            {#if d.editing && d.id > 0 && draftChanged(d)}
              →
              <select class="h-7 rounded border bg-background px-1.5 text-foreground" aria-label="버전 올리기" bind:value={d.bump}>
                <option value="minor">마이너 {nextVersion(d.version, false)}</option>
                <option value="major">메이저 {nextVersion(d.version, true)}</option>
              </select>
            {/if}
          </span>
          {#if d.source}
            <span>원본 <span class="font-mono text-foreground">{d.source}</span></span>
          {/if}
          <label class="flex items-center gap-1.5">
            모델
            {#if d.id === -1}
              <select class="h-7 max-w-72 rounded border bg-background px-1.5 text-foreground" value={chain.model} disabled={savingChain === chain.name} onchange={(e) => setModel(chain, e.currentTarget.value)}>
                {#each modelOptions(chain) as id (id)}
                  <option value={id}>{formatModelLabel(id)}</option>
                {/each}
              </select>
            {:else if d.editing}
              <select class="h-7 max-w-72 rounded border bg-background px-1.5 text-foreground" bind:value={d.model}>
                <option value="">체인 모델 ({formatModelLabel(chain.model)})</option>
                {#each modelOptions(chain) as id (id)}
                  <option value={id}>{formatModelLabel(id)}</option>
                {/each}
              </select>
            {:else}
              <span class="text-foreground">{d.model ? formatModelLabel(d.model) : `체인 모델 (${formatModelLabel(chain.model)})`}</span>
            {/if}
          </label>
        </div>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        {#if d.editing}
          <CButton variant="outline" size="small" disabled={savingDraft} onclick={() => void cancelEdit()}>취소</CButton>
          <CButton variant="default" size="small" disabled={!d.title.trim() || savingDraft || (d.id > 0 && !draftChanged(d))} onclick={() => void saveDraft()}>
            {savingDraft ? '저장 중...' : d.id === 0 ? '만들기' : '저장'}
          </CButton>
        {:else}
          {#if own}
            <CButton variant="outline" size="small" onclick={startEdit}>
              <ZIcon path={mdiPencil} class="mr-1" />
              편집
            </CButton>
          {/if}
          <CButton variant="outline" size="small" onclick={() => void openClone(d)}>
            <ZIcon path={mdiContentCopy} class="mr-1" />
            복제
          </CButton>
          {#if own && item}
            <CButton variant="outline" size="small" disabled={deletingPromptId === item.id} onclick={() => void delPrompt(item)}>
              <ZIcon path={mdiTrashCanOutline} class="mr-1" />
              삭제
            </CButton>
          {/if}
        {/if}
      </div>
    </div>
    {#if d.editing && d.id > 0 && draftChanged(d) && d.bump === 'major'}
      <div class="text-muted-foreground">메이저 범프: 이 틀로 돈 이전 결과와 다른 것으로 구분합니다(AI 편집은 다시 돌리는 대상이 없습니다).</div>
    {/if}
    {#if draftError}
      <div class="text-destructive">{draftError}</div>
    {/if}
    {#if d.editing && d.legacy}
      <div class="rounded border p-2 text-muted-foreground">예전 형식의 템플릿입니다. 저장하면 지금 내용이 system 지침이 되고, user 메시지는 사용자 입력에서 만들어집니다.</div>
    {/if}
    <!-- Each row pairs what is edited (left) with what it renders or fills (right): system, user, then the preview's inputs. -->
    <div class="grid items-start gap-3 md:grid-cols-2 md:items-stretch">
      <div class="relative min-w-0">
        <span class="pointer-events-none absolute -top-px right-[13px] rounded px-1.5 py-0.5 font-mono text-muted-foreground/70">system</span>
        <textarea
          id={`ai-template-system-${d.key}`}
          aria-label="system 지침"
          class="block h-96 w-full resize-y rounded border bg-background p-3 pr-16 font-mono text-[11px] text-foreground read-only:bg-muted/30 focus:outline-none focus:ring-1 focus:ring-ring read-only:focus:ring-0"
          readonly={!d.editing}
          placeholder="시스템 지침 및 규칙을 입력하세요."
          bind:value={d.messages[0].content}
        ></textarea>
      </div>
      <div class="relative min-w-0">
        <div class={renderedCell}>
        {#if specCheck?.ok && !specCheck.preview_error}
          {#each (specCheck.rendered_messages ?? []).filter((m) => m.role === 'system') as message, i (i)}
            {@render renderedMessage(message)}
          {/each}
        {/if}
        </div>
      </div>
      <fieldset class={`relative min-w-0 rounded border p-3 pr-16 text-xs md:min-h-48 ${d.editing ? 'bg-background' : 'bg-muted/30'}`}>
        <legend class="sr-only">user 메시지 변수</legend>
        <span class="pointer-events-none absolute -top-px right-[13px] rounded px-1.5 py-0.5 font-mono text-muted-foreground/70">user</span>
        <div class="grid items-start gap-3 sm:grid-cols-2 sm:gap-0 sm:divide-x">
          <div class="grid grid-cols-[auto_minmax(0,1fr)] items-start gap-x-2 gap-y-1 sm:pr-3">
            <div class="col-span-2 text-muted-foreground">예시 문서</div>
            {#each fixedVars(d.request_type) as f (f.name)}
              {@const off = !!f.after && !picked.includes(f.after)}
              <div class={`pt-1 font-mono ${off ? 'text-muted-foreground opacity-50' : ''}`} title={f.after ? `${f.after}를 고르면 들어갑니다` : '항상 들어갑니다'}>{f.name}</div>
              <div class={`min-w-0 ${off ? 'opacity-50' : ''}`}>{@render fixedField(f)}</div>
            {/each}
          </div>
          <div class="grid grid-cols-[auto_auto_minmax(0,1fr)] items-start gap-x-2 gap-y-1 sm:pl-3">
            <div class="col-span-3 text-muted-foreground">사용자 입력</div>
            {#each chain.spec.userInputs ?? [] as input (input.name)}
              {@const off = !picked.includes(input.name)}
              <input class="mt-1" type="checkbox" aria-label={`${input.name} 사용`} checked={!off} disabled={!d.editing} onchange={(e) => setPicked(d, chain, input.name, e.currentTarget.checked)} />
              <div class={`pt-1 font-mono ${off ? 'text-muted-foreground' : ''}`}>{input.name}</div>
              <div class="min-w-0">{@render inputField(input, off)}</div>
            {:else}
              <div class="col-span-3 pt-1 text-muted-foreground">없음</div>
            {/each}
          </div>
        </div>
      </fieldset>
      <div class="relative min-w-0">
        <div class={renderedCell}>
        {#if specCheck && !specCheck.ok}
          <div class="rounded border border-destructive/40 p-2 text-destructive">{specCheck.error}</div>
        {:else if specCheck?.preview_error}
          <div class="rounded border border-destructive/40 p-2 text-destructive">{specCheck.preview_error}</div>
        {:else if specCheck?.ok}
          {#each (specCheck.rendered_messages ?? []).filter((m) => m.role !== 'system') as message, i (i)}
            {@render renderedMessage(message)}
          {/each}
        {/if}
        </div>
      </div>
    </div>
  </div>
{/snippet}

{#snippet renderedMessage(message: ChainMessage)}
  <div class="relative flex flex-col overflow-hidden rounded border bg-muted/30 md:shrink-0 md:grow">
    <span class="pointer-events-none absolute -top-px right-[13px] rounded px-1.5 py-0.5 font-mono text-muted-foreground/70">{message.role}</span>
    <pre class="max-h-80 grow overflow-auto whitespace-pre-wrap p-3 pr-16 text-[11px] md:max-h-none">{message.content}</pre>
  </div>
{/snippet}

{#snippet runForm(inputs: ChainInput[])}
  <div class="rounded border bg-background p-3 text-xs text-foreground">
    <div class="mb-2 text-muted-foreground">사용자 입력</div>
    <div class="grid grid-cols-[auto_minmax(0,1fr)] items-start gap-x-4 gap-y-2">
      {#each inputs as input (input.name)}
        <div class="pt-1">{input.name}{input.required ? ' *' : ''}</div>
        <div class="min-w-0">{@render inputField(input, false)}</div>
      {:else}
        <div class="col-span-2 text-muted-foreground">없음</div>
      {/each}
    </div>
  </div>
{/snippet}

{#snippet inputField(input: ChainInput, disabled: boolean)}
  {@const label = humanize(input.name)}
  {@const given = previewValues[input.name] ?? input.default}
  {#if input.type === 'bool'}
    <input
      class="mt-1"
      {disabled}
      type="checkbox"
      aria-label={label}
      checked={given === true}
      onchange={(e) => setPreview(input.name, e.currentTarget.checked)}
    />
  {:else if input.type === 'textList'}
    <div class="space-y-1">
      {#each previewTexts(input.name) as item, i (i)}
        <div class="flex items-center gap-1">
          <textarea
            class="min-h-12 min-w-0 flex-1 rounded border bg-background px-2 py-1 disabled:bg-muted disabled:text-muted-foreground"
            {disabled}
            placeholder={input.placeholder ?? `${label} ${i + 1}`}
            value={item}
            oninput={(e) => setPreview(input.name, previewTexts(input.name).map((x, j) => (j === i ? e.currentTarget.value : x)))}
          ></textarea>
          <button
            type="button"
            class="rounded px-1.5 text-muted-foreground hover:bg-muted disabled:opacity-50"
            {disabled}
            aria-label={`${label} ${i + 1} 빼기`}
            onclick={() => setPreview(input.name, previewTexts(input.name).filter((_, j) => j !== i))}>✕</button
          >
        </div>
      {/each}
      <button
        type="button"
        class="rounded border px-2 py-0.5 text-muted-foreground hover:bg-muted disabled:opacity-50"
        {disabled}
        onclick={() => setPreview(input.name, [...previewTexts(input.name), ''])}>+ {label}</button
      >
    </div>
  {:else}
    {@const value = typeof given === 'string' ? given : ''}
    {#if input.type === 'choice'}
      <select class="h-7 w-full rounded border bg-background px-2 disabled:bg-muted disabled:text-muted-foreground" {disabled} {value} onchange={(e) => setPreview(input.name, e.currentTarget.value)}>
        {#if input.default === undefined}
          <option value="">{input.placeholder ?? '선택'}</option>
        {/if}
        {#each input.options ?? [] as option (option)}
          <option value={option}>{option}</option>
        {/each}
      </select>
    {:else if input.type === 'text'}
      <textarea
        class="min-h-12 w-full rounded border bg-background px-2 py-1 disabled:bg-muted disabled:text-muted-foreground"
        {disabled}
        placeholder={input.placeholder ?? ''}
        {value}
        oninput={(e) => setPreview(input.name, e.currentTarget.value)}
      ></textarea>
    {:else}
      <input
        class="h-7 w-full rounded border bg-background px-2 disabled:bg-muted disabled:text-muted-foreground"
        {disabled}
        placeholder={input.placeholder ?? ''}
        {value}
        oninput={(e) => setPreview(input.name, e.currentTarget.value)}
      />
    {/if}
  {/if}
{/snippet}

{#snippet fixedField(f: FixedVar)}
  <div class="whitespace-pre-wrap break-words pt-1 text-muted-foreground">{f.value}</div>
{/snippet}

{#snippet messagesView(messages: ChainMessage[])}
  {#each messages as m, i (i)}
    <div class="relative overflow-hidden rounded border bg-muted/30">
      <span class="pointer-events-none absolute -top-px right-[13px] rounded px-1.5 py-0.5 font-mono text-muted-foreground/70">{messageHeader(m)}</span>
      <pre class={`max-h-80 overflow-auto whitespace-pre-wrap p-3 text-[11px] ${m.if || m.each ? 'pt-6' : 'pr-16'}`}>{#each parts(m.content) as part, k (k)}{#if part.name !== undefined}<code
              class="var">{part.name}</code
            >{:else}{part.text}{/if}{/each}</pre>
    </div>
  {/each}
{/snippet}

{#snippet stepView(st: ChainStep, n: number, messages: ChainMessage[])}
  <div class="space-y-1.5 text-xs min-w-0">
    <div>
      <span class="text-muted-foreground">{n + 1}.</span>
      <span class="font-mono font-medium">{st.name}</span>
      {#if st.uses}
        <span class="ml-1 font-mono text-muted-foreground">{st.if ? `if ${st.if}: ` : ''}{st.inputs.filter((i) => i !== st.if).join(', ')} → {st.uses} → {st.output}</span>
      {:else}
        <span class="ml-1 font-mono text-muted-foreground">{st.inputs.join(', ')} → {st.format ?? 'text'}{st.output ? ` → ${st.output}` : ''}</span>
      {/if}
      <span class="ml-2 text-muted-foreground">{st.description}</span>
    </div>
    {@render messagesView(messages)}
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
                  {#if draft?.key === chain.name}
                    {@render templateEditor(chain, draft)}
                  {:else}
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
                    </div>
                    <div class="space-y-1 text-xs text-muted-foreground">
                      <div>버전: <span class="font-mono">{chain.metadata.version}</span></div>
                      {#if chain.spec.userInputs?.length}
                        {@render runForm(chain.spec.userInputs)}
                      {/if}
                      <div>task 입력: <span class="font-mono">{chain.spec.inputs.map(inputLabel).join(', ')}</span></div>
                    </div>
                    {#each chain.spec.steps as st, n (st.name)}
                      {@render stepView(st, n, st.messages)}
                    {/each}
                  </div>
                  {/if}
                </td>
              </tr>
            {/if}
          {/each}
          {#if draft?.key === 'new'}
            {@const chain = chainOf(draft.request_type)}
            {#if chain}
              <tr class="bg-blue-50 dark:bg-blue-950/40">
                <td class="p-3 whitespace-nowrap text-muted-foreground">{draft.owner}</td>
                <td class="p-3 whitespace-nowrap">{draft.title || '새 템플릿'}</td>
                <td class="p-3 text-muted-foreground" colspan="5">새 템플릿</td>
              </tr>
              <tr class="bg-muted/10">
                <td colspan="7" class="p-4">{@render templateEditor(chain, draft)}</td>
              </tr>
            {/if}
          {/if}
          {#each prompts as item (item.id)}
            {@const chain = chainOf(item.request_type)}
            {#if chain}
              {@const key = `prompt:${item.id}`}
              {@const itemExpanded = expanded === key}
              {@const own = item.user_id === $userInfo?.id}
              <tr
                class={['transition cursor-pointer', own ? 'bg-blue-50 hover:bg-blue-100 dark:bg-blue-950/40 dark:hover:bg-blue-900/40' : 'hover:bg-muted/30']}
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
              {#if itemExpanded && draft?.key === key}
                <tr class="bg-muted/10">
                  <td colspan="7" class="p-4">{@render templateEditor(chain, draft)}</td>
                </tr>
              {/if}
            {/if}
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>

<style>
  .var {
    padding: 0.125em 0.25em;
    border: 1px solid var(--color-border);
    border-radius: 0.25rem;
    background: light-dark(oklch(0.96 0.05 85), oklch(0.32 0.05 85));
    color: var(--color-code-foreground);
  }
</style>
