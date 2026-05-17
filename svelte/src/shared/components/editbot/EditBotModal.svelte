<svelte:options runes={true} />

<script lang="ts">
  import { mdiAutoFix, mdiCreation, mdiEye } from '@mdi/js'

  import mwapi from '$lib/utils/mwapi'
  import { showToast } from '$shared/ui/toast/toast'
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZButtonLink from '$shared/ui/ZButtonLink.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZModal from '$shared/ui/ZModal.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'
  import { getWikiViewHref } from '$shared/utils/wikiLink'

  interface Target {
    title: string
    storeUrl: string
    requestType: 'create' | 'edit'
    pageId?: number
    existingContent?: string
  }

  interface MwAllPagesResp {
    query?: {
      allpages?: Array<{ pageid: number; ns: number; title: string }>
    }
  }

  interface MwRevision {
    content?: string
    slots?: {
      main?: {
        content?: string
      }
    }
  }

  interface MwPage {
    title: string
    missing?: boolean
    revisions?: MwRevision[]
  }

  interface MwRawTextResp {
    query?: {
      pages?: MwPage[] | Record<string, MwPage>
    }
  }

  interface StoreResp {
    ok: boolean
    id: number
    created: boolean
  }

  let {
    show = false,
    target = null,
    onClose,
    onCreated,
  }: {
    show?: boolean
    target?: Target | null
    onClose?: () => void
    onCreated?: (data: StoreResp) => void
  } = $props()

  let promptList = $state<string[]>([])
  let promptTitle = $state('틀:프롬프트 생성')
  let llmPromptTemplate = $state('')
  let promptListLoading = $state(false)
  let promptContentLoading = $state(false)
  let existingContentLoading = $state(false)
  let submitting = $state(false)
  let llmInputDraft = $state('')
  let llmInputEdited = $state(false)
  let promptListFetched = false
  let wasShown = false
  let observedPromptTitle = ''
  let existingContent = $state('')

  let renderedLlmInput = $derived.by(() => renderFinalPrompt())
  let modalTitle = $derived(`'${target?.title ?? ''}' 문서 ${target?.requestType === 'edit' ? '편집' : '생성'}`)
  let modalTitleIconPath = $derived(target?.requestType === 'edit' ? mdiAutoFix : mdiCreation)
  let canSubmit = $derived(Boolean(target && llmInputDraft.trim() && !promptContentLoading && !existingContentLoading) && !submitting)

  $effect(() => {
    if (!show) {
      wasShown = false
      return
    }
    if (wasShown) return
    wasShown = true
    reset()
    if (target?.requestType === 'edit') {
      void fetchExistingContent(target.title)
    }
    if (!promptListFetched) {
      promptListFetched = true
      void fetchPromptList()
    }
  })

  $effect(() => {
    if (!show) return
    if (!promptTitle || promptTitle === observedPromptTitle) return
    observedPromptTitle = promptTitle
    void fetchPromptContent(promptTitle)
  })

  $effect(() => {
    if (!show) return
    if (llmInputEdited) return
    llmInputDraft = renderedLlmInput
  })

  function reset() {
    promptTitle = target?.requestType === 'edit' ? '틀:프롬프트 편집' : '틀:프롬프트 생성'
    llmPromptTemplate = ''
    promptContentLoading = false
    existingContentLoading = false
    existingContent = target?.existingContent ?? ''
    submitting = false
    llmInputEdited = false
    llmInputDraft = ''
    observedPromptTitle = ''
  }

  function asArray<T>(value: T[] | Record<string, T> | undefined): T[] {
    if (!value) return []
    return Array.isArray(value) ? value : Object.values(value)
  }

  function getRevisionContent(pageData: MwPage | undefined) {
    const revision = pageData?.revisions?.[0]
    return revision?.slots?.main?.content ?? revision?.content ?? ''
  }

  function renderTemplate(template: string, values: Record<string, string>) {
    let out = template
    for (const [key, value] of Object.entries(values)) {
      out = out.replaceAll(`{${key}}`, value)
    }
    return out
  }

  function renderFinalPrompt() {
    return renderTemplate(llmPromptTemplate, {
      제목: target?.title ?? '',
      기존문서: existingContent,
      추가요청: '',
      참고자료: '',
      분류: '',
    }).trimEnd()
  }

  async function fetchPromptList() {
    promptListLoading = true
    const [data, err] = await mwapi.get<MwAllPagesResp>({
      action: 'query',
      list: 'allpages',
      apprefix: '프롬프트',
      apnamespace: 10,
      aplimit: 'max',
    })
    promptListLoading = false

    if (err) {
      console.error(err)
      return
    }

    promptList = (data?.query?.allpages ?? []).map((p: { title: string }) => p.title)
    const preferred = target?.requestType === 'edit' ? '틀:프롬프트 편집' : '틀:프롬프트 생성'
    if (promptList.includes(preferred)) {
      promptTitle = preferred
      return
    }
    if (promptList.length > 0) {
      promptTitle = promptList[0]
    }
  }

  async function fetchExistingContent(title: string) {
    if (target?.existingContent != null) {
      existingContent = target.existingContent
      return
    }

    existingContentLoading = true
    const [data, err] = await mwapi.get<MwRawTextResp>({
      action: 'query',
      prop: 'revisions',
      rvprop: 'content',
      rvslots: 'main',
      titles: title,
    })
    existingContentLoading = false

    if (err) {
      console.error(err)
      showToast(err.message || '기존 문서를 불러오지 못했습니다.')
      return
    }

    const pageData = asArray<MwPage>(data?.query?.pages)[0]
    if (!pageData || pageData.missing) {
      showToast('기존 문서를 찾을 수 없습니다.')
      return
    }
    existingContent = getRevisionContent(pageData)
  }

  async function fetchPromptContent(title: string) {
    promptContentLoading = true
    llmPromptTemplate = ''
    const [data, err] = await mwapi.get<MwRawTextResp>({
      action: 'query',
      prop: 'revisions',
      rvprop: 'content',
      rvslots: 'main',
      titles: title,
    })
    promptContentLoading = false

    if (err) {
      console.error(err)
      showToast(err.message || '프롬프트를 불러오지 못했습니다.')
      return
    }

    const pageData = asArray<MwPage>(data?.query?.pages)[0]
    if (!pageData || pageData.missing) {
      showToast('프롬프트 문서를 찾을 수 없습니다.')
      return
    }

    llmPromptTemplate = getRevisionContent(pageData)
  }

  async function ok() {
    if (!target || !canSubmit) return

    submitting = true
    const [data, err] = await httpy.post<StoreResp>(target.storeUrl, {
      page_id: target.pageId,
      prompt_title: promptTitle,
      request_type: target.requestType,
      llm_input: llmInputDraft,
    })
    submitting = false

    if (err) {
      console.error(err)
      showToast(err.message || '편집봇 등록 실패')
      return
    }
    if (!data) return

    showToast(data.created ? '편집봇에 등록했습니다.' : '이미 편집봇에 등록되어 있습니다.', {
      timeout: 8000,
      placement: 'center',
      action: {
        label: '편집봇 작업목록 바로가기',
        href: '/tool/editbot',
      },
    })
    onCreated?.(data)
    onClose?.()
  }

  function onChangeLlmInput(e: Event) {
    llmInputEdited = true
    llmInputDraft = (e.currentTarget as HTMLTextAreaElement).value
  }

  function resetLlmInputToRendered() {
    llmInputEdited = false
    llmInputDraft = renderedLlmInput
  }

  function cancel() {
    onClose?.()
  }
</script>

<ZModal
  {show}
  title={modalTitle}
  titleIconPath={modalTitleIconPath}
  okText={submitting ? '등록 중' : '등록'}
  okColor="primary"
  okDisabled={!canSubmit}
  backdropClosable={true}
  panelClass="h-[80vh] w-[90vw] md:w-[75vw]"
  on:ok={ok}
  on:cancel={cancel}
>
  <div class="flex h-full min-h-0 flex-col gap-4 overflow-y-auto">
    <div class="flex items-center gap-2">
      <span class="shrink-0 text-sm text-(--color-subtle)">템플릿</span>
      {#if promptListLoading}
        <div class="flex h-9 items-center">
          <ZSpinner />
        </div>
      {:else}
        <select class="z-select min-w-0 flex-1" bind:value={promptTitle}>
          {#each promptList as title (title)}
            <option value={title}>{title}</option>
          {/each}
          {#if promptList.length === 0}
            <option value="틀:프롬프트 생성">틀:프롬프트 생성</option>
          {/if}
        </select>
        <ZButtonLink
          color="default"
          href={getWikiViewHref(promptTitle)}
          rel="external noopener noreferrer"
          target="_blank"
          title={`${promptTitle} 보기`}
          data-sveltekit-reload
        >
          <ZIcon path={mdiEye} />
        </ZButtonLink>
      {/if}
    </div>

    <div class="grid min-h-0 flex-1 grid-rows-[auto_minmax(0,1fr)] gap-1">
      <div class="flex items-center gap-2 text-sm text-(--color-subtle)">
        <label for="editbot-llm-input">프롬프트</label>
        {#if llmInputEdited}
          <ZButton size="small" color="default" onclick={resetLlmInputToRendered}>리셋</ZButton>
        {/if}
      </div>
      {#if promptContentLoading}
        <div class="flex min-h-64 items-center justify-center rounded border border-(--border-color-subtle) md:min-h-0">
          <ZSpinner />
        </div>
      {:else}
        <textarea
          id="editbot-llm-input"
          class="z-input h-full min-h-0 resize-none font-mono text-xs"
          value={llmInputDraft}
          oninput={onChangeLlmInput}
          placeholder="템플릿이 로딩되면 렌더링된 프롬프트가 표시됩니다."
        ></textarea>
      {/if}
    </div>
  </div>
</ZModal>
