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
  }

  // A step's messages as written, a jinja2 template per role (prompts.MessageTemplates).
  interface MessageTemplates {
    system?: string
    user?: string
  }

  const messageRoles = ['system', 'user'] as const

  interface PromptEditorState {
    name: string
    description: string
    message: Required<MessageTemplates>
    input_variables: string[]
  }

  // A step: a jinja2 template per message, read in its output variable's format (prompts.Step).
  interface ChainStep {
    name: string
    description?: string
    message: MessageTemplates
    output_variable?: { name?: string; format?: string }
    inputs: string[]
  }

  // A variable (prompts.Input); type is read by the server from json_schema and ui.
  interface ChainInput {
    name: string
    description?: string
    type: 'string' | 'text' | 'textList' | 'bool' | 'choice'
    json_schema?: { type: string; enum?: string[] }
    ui?: { variant?: string; placeholder?: string }
    is_required?: boolean
    default?: string | boolean
  }

  interface Chain {
    name: string
    description: string
    version: string
    user_inputs?: ChainInput[]
    inputs: ChainInput[]
    steps: ChainStep[]
    model: string
    uses: number
    used_at?: string | null
    // The chain as an AI edit prompt template in YAML, which a clone starts from.
    template: string
  }

  interface PromptItem {
    id: number
    user_id: number
    user_name: string
    title: string
    description: string
    request_type: string
    use_count: number
    spec: string
    user_inputs: ChainInput[]
    error?: string
    version: string
    model: string
    used_at?: string | null
    updated_at?: string
  }

  // A user input's name as the run form shows it: someName as "Some name" (as in the skin's AiEditRunner).
  function humanize(name: string) {
    const words = name.replace(/([a-z0-9])([A-Z])/g, '$1 $2').toLowerCase()
    return words.charAt(0).toUpperCase() + words.slice(1)
  }

  type PreviewValue = string | boolean | string[]

  // Values typed into a preview, by name: the run form's and the task's (title, body). Kept on screen only; one
  // left out takes the server's sample.
  let previewValues = $state<Record<string, PreviewValue>>({})

  function previewTexts(name: string): string[] {
    const v = previewValues[name]
    return Array.isArray(v) ? v : []
  }

  function setPreview(name: string, v: PreviewValue) {
    previewValues = { ...previewValues, [name]: v }
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
    chains = (data ?? [])
      .map((c, i) => ({ c, i }))
      .sort((a, b) => usedAt(b.c) - usedAt(a.c) || a.i - b.i)
      .map(({ c }) => c)
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
  // for a new one (a clone) and -1 for a built-in chain, which is only seen. spec is the template's YAML; model,
  // outside it, is the model it runs on ('' for its chain's).
  interface TemplateDraft {
    key: string
    editing: boolean
    id: number
    owner: string
    source: string
    request_type: string
    spec: string
    model: string
    version: string
    original: string
    bump: 'minor' | 'major'
  }

  // nextVersion is prompts.NextVersion: the minor, or with major the major with the minor back to 0.
  function nextVersion(version: string, major: boolean) {
    const [m, n] = version
      .replace(/^v/, '')
      .split('.')
      .map((x) => Number(x) || 0)
    return major ? `v${m + 1}.0` : `v${m}.${(n ?? 0) + 1}`
  }

  let draft = $state<TemplateDraft | null>(null)
  let draftError = $state('')
  let savingDraft = $state(false)

  function draftState(d: Pick<TemplateDraft, 'spec' | 'model'>) {
    return JSON.stringify([d.spec, d.model])
  }

  function draftChanged(d: TemplateDraft) {
    return draftState(d) !== d.original
  }

  // The version moves with the spec only; a change of model does not bump it.
  function specChanged(d: TemplateDraft) {
    return JSON.parse(d.original)[0] !== d.spec
  }

  function newDraft(d: Omit<TemplateDraft, 'original' | 'bump'>): TemplateDraft {
    return { ...d, original: draftState(d), bump: 'minor' }
  }

  function promptDraft(item: PromptItem, chain: Chain): TemplateDraft {
    return newDraft({
      key: `prompt:${item.id}`,
      editing: false,
      id: item.id,
      owner: item.user_name,
      source: chain.name,
      request_type: item.request_type,
      spec: item.spec ?? '',
      model: item.model ?? '',
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
      request_type: requestType,
      spec: chain.template,
      model: chain.model,
      version: chain.version,
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
    if (savingDraft || !(await confirmDiscard())) return
    draftError = ''
    expanded = expanded === key ? null : key
    draft = expanded ? draftFor(expanded) : null
    promptEditor = null
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
    promptEditor = null
  }

  // A clone's name: its source's with the cloning user's name in lower case (doc-update-jmnote).
  function cloneName(name: string) {
    const user = ($userInfo?.name ?? '').toLowerCase()
    return user ? `${name}-${user}` : name
  }

  // The spec's top-level name, as YAML writes it plainly or quoted.
  function specName(spec: string) {
    const m = spec.match(/^name:\s*(.*)$/m)
    if (!m) return ''
    const v = m[1].trim()
    try {
      return v.startsWith('"') ? JSON.parse(v) : v.replace(/^'(.*)'$/, '$1')
    } catch {
      return v
    }
  }

  // The spec with its name replaced; a JSON string is a YAML double-quoted scalar.
  function renamed(spec: string, name: string) {
    return spec.replace(/^name:.*$/m, `name: ${JSON.stringify(name)}`)
  }

  // A clone opens as a new template being edited, at the top of the templates.
  async function openClone(source: TemplateDraft) {
    if (savingDraft || !(await confirmDiscard())) return
    draftError = ''
    const chain = chainOf(source.request_type)
    draft = {
      ...source,
      key: 'new',
      editing: true,
      id: 0,
      owner: $userInfo?.name ?? '',
      source: chain?.name ?? '',
      version: 'v1.0',
      spec: renamed(source.spec, cloneName(specName(source.spec))),
      // A clone takes the model its source runs on now as a value of its own, so later changes do not follow.
      model: source.model || chain?.model || '',
      original: '',
      bump: 'minor',
    }
    promptEditor = null
    expanded = 'new'
  }

  interface SpecCheck {
    ok: boolean
    error?: string
    user_inputs?: ChainInput[]
    editor?: {
      name: string
      description: string
      message: MessageTemplates
      input_variables: Array<{ name: string }>
    }
    rendered_messages?: ChainMessage[]
    preview_error?: string
  }

  // The spec as the server reads it on save (POST /api/ai-prompts/check), shortly after typing stops.
  let specCheck = $state<SpecCheck | null>(null)
  let promptEditor = $state<PromptEditorState | null>(null)

  function quoteYaml(value: string) {
    return JSON.stringify(value)
  }

  // A message as a YAML block under message: its lines indented by four, with the indentation given when its first
  // line starts with a space (YAML would read it from that line otherwise).
  function messageBlock(role: string, content: string) {
    const indicator = /^(\s*\n)*[ \t]/.test(content) ? '|2-' : '|-'
    const lines = content.split('\n').map((line) => (line ? `    ${line}` : ''))
    return `  ${role}: ${indicator}\n${lines.join('\n')}`
  }

  function writeEditorSpec(editor: PromptEditorState) {
    const blocks = messageRoles.filter((role) => editor.message[role].trim()).map((role) => messageBlock(role, editor.message[role]))
    const message = blocks.length ? `message:\n${blocks.join('\n')}` : 'message: {}'
    const inputVariables = editor.input_variables.length
      ? editor.input_variables.map((name) => `  - name: ${quoteYaml(name)}`).join('\n')
      : '[]'
    const description = editor.description.trim() ? `description: ${quoteYaml(editor.description)}\n` : ''
    const variables = inputVariables === '[]' ? ' []' : `\n${inputVariables}`
    return `name: ${quoteYaml(editor.name)}\n${description}${message}\ninput_variables:${variables}\n`
  }

  function updatePromptEditor(patch: Partial<PromptEditorState>) {
    if (!draft || !promptEditor) return
    promptEditor = { ...promptEditor, ...patch }
    draft.spec = writeEditorSpec(promptEditor)
  }

  function editorInputs(chain: Chain): ChainInput[] {
    const all = [...chain.inputs, ...(chain.user_inputs ?? [])]
    return all.filter((input, index) => all.findIndex((candidate) => candidate.name === input.name) === index)
  }

  $effect(() => {
    const d = draft
    if (!d) {
      specCheck = null
      return
    }
    const spec = d.spec
    const requestType = d.request_type
    const values = $state.snapshot(previewValues)
    let active = true
    const timer = setTimeout(async () => {
      const [data] = await httpy.post<SpecCheck>('/api/ai-prompts/check', { request_type: requestType, spec, preview_values: values })
      if (active && data) {
        specCheck = data
        if (!promptEditor && data.ok && data.editor) {
          promptEditor = {
            name: data.editor.name,
            description: data.editor.description,
            message: { system: data.editor.message.system ?? '', user: data.editor.message.user ?? '' },
            input_variables: data.editor.input_variables.map((input) => input.name),
          }
        }
      }
    }, 400)
    return () => {
      active = false
      clearTimeout(timer)
    }
  })

  async function saveDraft() {
    if (!draft) return
    savingDraft = true
    const submitted = draft
    const [data, err] = await httpy.post<PromptItem>('/api/ai-prompts', {
      id: draft.id,
      request_type: draft.request_type,
      spec: draft.spec,
      model: draft.model,
      bump: draft.bump,
    })
    if (err || !data) {
      savingDraft = false
      draftError = err?.message || '저장 실패'
      return
    }
    showToast(submitted.id === 0 ? '복제했습니다' : '저장했습니다')
    await fetchPrompts()
    savingDraft = false
    // Rows stay as they are while saving (toggle and clone wait), so the saved template opens as it now is.
    if (draft === submitted) {
      expanded = `prompt:${data.id}`
      draft = draftFor(expanded)
      promptEditor = null
    }
  }

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
  <div class="space-y-3 text-xs">
    <div class="flex flex-wrap items-start gap-3">
      <div class="flex min-w-0 flex-1 flex-wrap items-center gap-x-4 gap-y-1 text-muted-foreground">
        {#if d.owner}
          <span>소유자 <span class="text-foreground">{d.owner}</span></span>
        {/if}
        <span class="flex items-center gap-1.5">
          버전 <span class="font-mono text-foreground">{d.version}</span>
          {#if d.editing && d.id > 0 && specChanged(d)}
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
        {#if d.id === -1}
          <label class="flex items-center gap-1.5">
            모델
            <select
              class="h-7 max-w-72 rounded border bg-background px-1.5 text-foreground"
              value={chain.model}
              disabled={savingChain === chain.name}
              onchange={(e) => setModel(chain, e.currentTarget.value)}
            >
              {#each modelOptions(chain) as id (id)}
                <option value={id}>{formatModelLabel(id)}</option>
              {/each}
            </select>
          </label>
        {:else if d.editing}
          <label class="flex items-center gap-1.5">
            모델
            <select class="h-7 max-w-72 rounded border bg-background px-1.5 text-foreground" bind:value={d.model}>
              <option value="">체인 모델 ({formatModelLabel(chain.model)})</option>
              {#each modelOptions(chain) as id (id)}
                <option value={id}>{formatModelLabel(id)}</option>
              {/each}
            </select>
          </label>
        {:else}
          <span
            >모델 <span class="text-foreground">{d.model ? formatModelLabel(d.model) : `체인 모델 (${formatModelLabel(chain.model)})`}</span
            ></span
          >
        {/if}
      </div>
      <div class="flex flex-wrap items-center gap-2">
        {#if d.editing}
          <CButton variant="outline" size="small" disabled={savingDraft} onclick={() => void cancelEdit()}>취소</CButton>
          <CButton
            variant="default"
            size="small"
            disabled={savingDraft || !specCheck?.ok || (d.id > 0 && !draftChanged(d))}
            onclick={() => void saveDraft()}
          >
            {savingDraft ? '저장 중...' : d.id === 0 ? '만들기' : '저장'}
          </CButton>
        {:else}
          {#if own}
            <CButton variant="outline" size="small" disabled={!promptEditor} onclick={startEdit}>
              <ZIcon path={mdiPencil} class="mr-1" />
              편집
            </CButton>
          {/if}
          <CButton variant="outline" size="small" disabled={!promptEditor} onclick={() => void openClone(d)}>
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
    {#if d.editing && d.id > 0 && specChanged(d) && d.bump === 'major'}
      <div class="text-muted-foreground">
        메이저 범프: 이 틀로 돈 이전 결과와 다른 것으로 구분합니다(AI 편집은 다시 돌리는 대상이 없습니다).
      </div>
    {/if}
    {#if draftError}
      <div class="text-destructive">{draftError}</div>
    {/if}
    <div class="grid items-start gap-3 md:grid-cols-2 md:items-stretch">
      <div class="min-w-0 space-y-3">
        {#if promptEditor && !d.editing}
          <div class="space-y-3">
            <div class="space-y-1">
              <div class="font-medium">{promptEditor.name}</div>
              {#if promptEditor.description}
                <div class="text-muted-foreground">{promptEditor.description}</div>
              {/if}
            </div>
            <div class="flex flex-wrap items-center gap-1.5">
              <span class="text-muted-foreground">입력값</span>
              {#each editorInputs(chain).filter((input) => promptEditor?.input_variables.includes(input.name)) as input (input.name)}
                <code class="rounded bg-muted px-1.5 py-0.5" title={input.description}>{input.name}</code>
              {/each}
            </div>
            {#each messageRoles.filter((role) => promptEditor?.message[role].trim()) as role (role)}
              <div class="relative overflow-hidden rounded border bg-muted/30">
                <span class="pointer-events-none absolute -top-px right-[13px] rounded px-1.5 py-0.5 font-mono text-muted-foreground/70"
                  >{role}</span
                >
                <pre class="max-h-[36rem] overflow-auto whitespace-pre-wrap p-3 pr-24 font-mono text-xs leading-relaxed">{promptEditor
                    .message[role]}</pre>
              </div>
            {/each}
          </div>
        {:else if promptEditor}
          <div class="space-y-3">
            <div class="grid gap-2 sm:grid-cols-2">
              <label class="space-y-1">
                <span class="block text-muted-foreground">템플릿 이름</span>
                <input
                  class="h-8 w-full rounded border bg-background px-2"
                  value={promptEditor.name}
                  oninput={(e) => updatePromptEditor({ name: e.currentTarget.value })}
                />
              </label>
              <label class="space-y-1">
                <span class="block text-muted-foreground">설명</span>
                <input
                  class="h-8 w-full rounded border bg-background px-2"
                  value={promptEditor.description}
                  oninput={(e) => updatePromptEditor({ description: e.currentTarget.value })}
                />
              </label>
            </div>
            <fieldset class="rounded border bg-background p-3">
              <legend class="px-1 text-muted-foreground">사용할 입력값</legend>
              <p class="mb-2 text-muted-foreground">작업 입력값은 필수입니다. 사용자 입력값은 본문에서 쓸 때 선택하세요.</p>
              <div class="flex flex-wrap gap-x-4 gap-y-2">
                {#each editorInputs(chain) as input (input.name)}
                  <label class="flex items-center gap-1.5" title={input.description}>
                    <input
                      type="checkbox"
                      checked={promptEditor.input_variables.includes(input.name)}
                      disabled={chain.inputs.some((required) => required.name === input.name)}
                      onchange={(e) => {
                        const selected = new Set(promptEditor?.input_variables ?? [])
                        if (e.currentTarget.checked) selected.add(input.name)
                        else selected.delete(input.name)
                        updatePromptEditor({ input_variables: [...selected] })
                      }}
                    />
                    <span>{humanize(input.name)}</span>
                    <code class="text-muted-foreground">{input.name}</code>
                  </label>
                {/each}
              </div>
            </fieldset>
            {#each messageRoles as role (role)}
              <section class="overflow-hidden rounded border bg-background">
                <header class="border-b bg-muted/30 px-3 py-2 font-mono text-muted-foreground">{role}</header>
                <textarea
                  class="min-h-64 w-full resize-y bg-background p-3 font-mono text-xs leading-relaxed focus:outline-none"
                  aria-label={`${role} 메시지 내용`}
                  value={promptEditor.message[role]}
                  oninput={(e) => {
                    if (promptEditor) updatePromptEditor({ message: { ...promptEditor.message, [role]: e.currentTarget.value } })
                  }}
                ></textarea>
              </section>
            {/each}
            <details class="rounded border px-3 py-2 text-muted-foreground">
              <summary class="cursor-pointer">본문에서 변수와 조건을 쓰는 방법</summary>
              <div class="mt-2 space-y-1">
                <div><code>{'{{ title }}'}</code>처럼 입력값을 넣습니다.</div>
                <div><code>{'{% if backlinks %}…{% endif %}'}</code>처럼 조건을 붙입니다.</div>
                <div>
                  대문자와 숫자로 된 MediaWiki 변수(예: <code>{'{{CURRENTDAY2}}'}</code>)와 콜론형 위키 문법은 그대로 둡니다. 알려지지 않은
                  소문자 이름을 글자로 남길 때는 <code>{'{% raw %}…{% endraw %}'}</code>를 씁니다.
                </div>
              </div>
            </details>
          </div>
        {:else if specCheck && !specCheck.ok}
          <pre class="max-h-[36rem] overflow-auto whitespace-pre-wrap rounded border bg-muted/30 p-3 font-mono text-[11px]">{d.spec}</pre>
        {:else}
          <div class="rounded border p-3 text-muted-foreground">작성 화면을 준비하고 있습니다…</div>
        {/if}
      </div>
      <div class="relative min-w-0 space-y-2">
        {#if specCheck && !specCheck.ok}
          <div class="rounded border border-destructive/40 p-2 whitespace-pre-wrap text-destructive">{specCheck.error}</div>
        {:else if specCheck?.ok}
          {@render previewForm(chain, specCheck.user_inputs ?? [])}
          {#if specCheck.preview_error}
            <div class="rounded border border-destructive/40 p-2 text-destructive">{specCheck.preview_error}</div>
          {:else}
            {#each specCheck.rendered_messages ?? [] as message, i (i)}
              <div class="relative overflow-hidden rounded border bg-muted/30">
                <span class="pointer-events-none absolute -top-px right-[13px] rounded px-1.5 py-0.5 font-mono text-muted-foreground/70"
                  >{message.role}</span
                >
                <pre class="max-h-80 overflow-auto whitespace-pre-wrap p-3 pr-16 text-[11px]">{message.content}</pre>
              </div>
            {/each}
          {/if}
        {/if}
      </div>
    </div>
  </div>
{/snippet}

{#snippet previewForm(chain: Chain, inputs: ChainInput[])}
  <fieldset class="rounded border bg-background p-3">
    <legend class="px-1 text-muted-foreground">미리보기 값 (저장되지 않음)</legend>
    <div class="grid grid-cols-[auto_minmax(0,1fr)] items-start gap-x-3 gap-y-1.5">
      {#each chain.inputs as input (input.name)}
        <div class="pt-1 font-mono" title="작업이 채우는 값">{input.name}</div>
        <textarea
          class="min-h-8 w-full rounded border bg-background px-2 py-1"
          placeholder={input.name === 'title' ? 'Hello World' : '문서 시작했다\n문서 끝났다'}
          value={typeof previewValues[input.name] === 'string' ? (previewValues[input.name] as string) : ''}
          oninput={(e) => setPreview(input.name, e.currentTarget.value)}
        ></textarea>
      {/each}
      {#each inputs as input (input.name)}
        <div class="pt-1 font-mono" title={input.description}>{input.name}{input.is_required ? ' *' : ''}</div>
        <div class="min-w-0">{@render inputField(input)}</div>
      {/each}
    </div>
  </fieldset>
{/snippet}

{#snippet inputField(input: ChainInput)}
  {@const label = humanize(input.name)}
  {@const given = previewValues[input.name] ?? input.default}
  {@const placeholder = input.ui?.placeholder ?? ''}
  {#if input.type === 'bool'}
    <input
      class="mt-1"
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
            class="min-h-8 min-w-0 flex-1 rounded border bg-background px-2 py-1"
            placeholder={placeholder || `${label} ${i + 1}`}
            value={item}
            oninput={(e) =>
              setPreview(
                input.name,
                previewTexts(input.name).map((x, j) => (j === i ? e.currentTarget.value : x)),
              )}
          ></textarea>
          <button
            type="button"
            class="rounded px-1.5 text-muted-foreground hover:bg-muted"
            aria-label={`${label} ${i + 1} 빼기`}
            onclick={() =>
              setPreview(
                input.name,
                previewTexts(input.name).filter((_, j) => j !== i),
              )}>✕</button
          >
        </div>
      {/each}
      <button
        type="button"
        class="rounded border px-2 py-0.5 text-muted-foreground hover:bg-muted"
        onclick={() => setPreview(input.name, [...previewTexts(input.name), ''])}>+ {label}</button
      >
    </div>
  {:else}
    {@const value = typeof given === 'string' ? given : ''}
    {#if input.type === 'choice'}
      <select class="h-7 w-full rounded border bg-background px-2" {value} onchange={(e) => setPreview(input.name, e.currentTarget.value)}>
        {#if input.default === undefined}
          <option value="">{placeholder || '선택'}</option>
        {/if}
        {#each input.json_schema?.enum ?? [] as option (option)}
          <option value={option}>{option}</option>
        {/each}
      </select>
    {:else if input.type === 'text'}
      <textarea
        class="min-h-12 w-full rounded border bg-background px-2 py-1"
        {placeholder}
        {value}
        oninput={(e) => setPreview(input.name, e.currentTarget.value)}
      ></textarea>
    {:else}
      <input
        class="h-7 w-full rounded border bg-background px-2"
        {placeholder}
        {value}
        oninput={(e) => setPreview(input.name, e.currentTarget.value)}
      />
    {/if}
  {/if}
{/snippet}

{#snippet stepView(st: ChainStep, n: number)}
  <div class="space-y-1.5 text-xs min-w-0">
    <div>
      <span class="text-muted-foreground">{n + 1}.</span>
      <span class="font-mono font-medium">{st.name}</span>
      <span class="ml-1 font-mono text-muted-foreground"
        >{st.inputs.join(', ')} → {st.output_variable?.format ?? 'text'}{st.output_variable?.name
          ? ` → ${st.output_variable.name}`
          : ''}</span
      >
      <span class="ml-2 text-muted-foreground">{st.description}</span>
    </div>
    {#each messageRoles.filter((role) => st.message[role]) as role (role)}
      <div class="relative overflow-hidden rounded border bg-muted/30">
        <span class="pointer-events-none absolute -top-px right-[13px] rounded px-1.5 py-0.5 font-mono text-muted-foreground/70"
          >{role}</span
        >
        <pre class="max-h-96 overflow-auto whitespace-pre-wrap p-3 pr-16 text-[11px]">{st.message[role]}</pre>
      </div>
    {/each}
  </div>
{/snippet}

{#snippet promptRows(items: PromptItem[])}
  {#each items as item (item.id)}
    {@const chain = chainOf(item.request_type)}
    {#if chain}
      {@const key = `prompt:${item.id}`}
      {@const itemExpanded = expanded === key}
      {@const own = item.user_id === $userInfo?.id}
      <tr
        class={[
          'transition cursor-pointer',
          own ? 'bg-blue-50 hover:bg-blue-100 dark:bg-blue-950/40 dark:hover:bg-blue-900/40' : 'hover:bg-muted/30',
        ]}
        onclick={() => toggle(key)}
      >
        <td class="p-3 whitespace-nowrap text-muted-foreground">{item.user_name}</td>
        <td class="p-3 whitespace-nowrap"
          >{item.title}{#if item.error}<span class="ml-1 text-destructive" title={item.error}>⚠</span>{/if}</td
        >
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
          {#if draft?.key === 'new'}
            {@const chain = chainOf(draft.request_type)}
            {#if chain}
              <tr class="bg-blue-50 dark:bg-blue-950/40">
                <td class="p-3 whitespace-nowrap text-muted-foreground">{draft.owner}</td>
                <td class="p-3 whitespace-nowrap">{specName(draft.spec) || '새 템플릿'}</td>
                <td class="p-3 text-muted-foreground" colspan="5">새 템플릿</td>
              </tr>
              <tr class="bg-muted/10">
                <td colspan="7" class="p-4">{@render templateEditor(chain, draft)}</td>
              </tr>
            {/if}
          {/if}
          {@render promptRows(prompts.filter((item) => item.user_id === $userInfo?.id))}
          {#each chains as chain (chain.name)}
            {@const isExpanded = expanded === chain.name}
            <tr class="bg-muted/50 hover:bg-muted transition cursor-pointer" onclick={() => toggle(chain.name)}>
              <td class="p-3 text-muted-foreground">-</td>
              <td class="p-3 whitespace-nowrap">{chain.name}</td>
              <td class="p-3 text-muted-foreground">{chain.description}</td>
              <td class="p-3 text-xs">{formatModelLabel(chain.model)}</td>
              <td class="p-3 text-center text-muted-foreground tabular-nums" title="이 체인이 끝까지 돈 횟수(프롬프트 틀로 돈 것 포함)"
                >{chain.uses || '-'}</td
              >
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
                        <div>버전: <span class="font-mono">{chain.version}</span></div>
                        <div>task 입력: <span class="font-mono">{chain.inputs.map((i) => i.name).join(', ')}</span></div>
                      </div>
                      {#each chain.steps as st, n (st.name)}
                        {@render stepView(st, n)}
                      {/each}
                    </div>
                  {/if}
                </td>
              </tr>
            {/if}
          {/each}
          {@render promptRows(prompts.filter((item) => item.user_id !== $userInfo?.id))}
        </tbody>
      </table>
    </div>
  </div>
</div>
