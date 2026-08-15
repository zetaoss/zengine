<svelte:options runes={true} />

<script lang="ts">
  import { mdiChevronDown, mdiChevronUp, mdiSend } from '@mdi/js'
  import { onDestroy } from 'svelte'

  import CButton from '$shared/ui/CButton.svelte'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSelect from '$shared/ui/ZSelect.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import ZTextarea from '$shared/ui/ZTextarea.svelte'
  import httpy from '$shared/utils/httpy'

  import { replaceWikiEditorContent } from '../wikiEditor'
  import type { AIEditRequestType, AIEditTaskResult } from './aiEditTypes'
  import { selectedAiEditResult } from './selectedAiEditResult'

  let {
    pageId,
    title,
    requestType,
    submittedTaskId,
    resetToken,
    onPollingChange,
    onPhaseChange,
    onRequestTypeChange,
  }: {
    pageId: number | undefined
    title: string
    requestType: AIEditRequestType
    submittedTaskId: number | undefined
    resetToken: number
    onPollingChange: (polling: boolean) => void
    onPhaseChange: (phase: AIEditTaskResult['phase'] | undefined) => void
    onRequestTypeChange?: (newType: AIEditRequestType) => void
  } = $props()

  const pollingInitialMs = 1000
  const pollingBackoffFactor = 1.1
  const recentResultWindowMs = 24 * 60 * 60 * 1000
  const activeTaskPhases: ReadonlySet<AIEditTaskResult['phase']> = new Set(['Pending', 'Generating', 'Retrying'])

  let tasks = $state<AIEditTaskResult[]>([])
  let selectedTaskId = $state('')
  let resultOutput = $state('')
  let tasksLoading = $state(false)
  let initialLoad = $state(true)
  let resultsExpanded = $state(false)
  let resultExpansionInitialized = $state(false)
  let handledSubmittedTaskId: number | undefined
  let handledResetToken: number | undefined
  let pollTimer: ReturnType<typeof setTimeout> | undefined
  let pollDelay = pollingInitialMs
  let pollSerial = 0

  interface TaskSelectItem {
    value: string
    label: string
    secondaryLabel?: string
    dateLabel?: string
  }

  function formatOutputLength(output: string | null) {
    const characters = Array.from(output ?? '').length
    return `${characters.toLocaleString('ko-KR')}자`
  }

  function formatTaskDate(value: string) {
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return ''

    const now = new Date()
    if (date.toDateString() === now.toDateString()) {
      return `${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
    }

    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
  }

  let resultTasks = $derived(
    tasks.filter((task) => task.phase === 'Completed' && task.llm_output?.trim().length),
  )
  let hasRecentResults = $derived(
    resultTasks.some((task) => {
      const createdAt = Date.parse(task.created_at)
      return Number.isFinite(createdAt) && Date.now() - createdAt <= recentResultWindowMs
    }),
  )
  let taskSelectItems = $derived(
    resultTasks.map((task) => {
      const dateLabel = formatTaskDate(task.created_at)
      return {
        value: String(task.id),
        label: `#${task.id}`,
        secondaryLabel: formatOutputLength(task.llm_output),
        dateLabel,
      }
    }),
  )

  $effect(() => {
    void fetchTasks()
  })

  $effect(() => {
    if (resultExpansionInitialized || resultTasks.length === 0) return
    resultsExpanded = hasRecentResults
    resultExpansionInitialized = true
  })

  $effect(() => {
    if (handledResetToken === undefined) {
      handledResetToken = resetToken
      return
    }
    if (resetToken === handledResetToken) return
    handledResetToken = resetToken
    handledSubmittedTaskId = undefined
    onPhaseChange(undefined)
    stopPolling()
  })

  $effect(() => {
    if (submittedTaskId === undefined || submittedTaskId === handledSubmittedTaskId) return
    handledSubmittedTaskId = submittedTaskId
    stopPolling()
    pollDelay = pollingInitialMs
    void fetchResultTask(submittedTaskId, ++pollSerial)
  })

  $effect(() => {
    if (!selectedTaskId) {
      resultOutput = ''
      return
    }
    resultOutput = resultTasks.find((task) => String(task.id) === selectedTaskId)?.llm_output ?? ''
  })

  $effect(() => {
    const task = resultTasks.find((item) => String(item.id) === selectedTaskId)
    selectedAiEditResult.set(task ? { content: resultOutput, taskId: task.id, requestType, onRequestTypeChange } : null)
  })

  onDestroy(() => {
    stopPolling()
    selectedAiEditResult.set(null)
  })

  function setPolling(value: boolean) {
    onPollingChange(value)
  }

  function stopPolling() {
    setPolling(false)
    pollSerial += 1
    if (pollTimer !== undefined) {
      clearTimeout(pollTimer)
      pollTimer = undefined
    }
  }

  async function fetchResultTask(taskId: number, serial: number) {
    if (serial !== pollSerial) return
    console.log(`[AIEdit] Polling task ${taskId} (delay: ${pollDelay}ms)`)
    setPolling(true)

    const [data, err] = await httpy.get<AIEditTaskResult>(`/api/ai-edit/${taskId}`)
    if (serial !== pollSerial) return
    if (err || !data) {
      setPolling(false)
      return
    }

    onPhaseChange(data.phase)
    resultOutput = data.llm_output ?? ''
    if (!activeTaskPhases.has(data.phase)) {
      setPolling(false)
      await fetchTasks()
      if (!tasks.some((task) => task.id === taskId)) tasks = [data, ...tasks]
      selectedTaskId = String(taskId)
      return
    }

    pollTimer = setTimeout(() => {
      pollDelay = Math.ceil(pollDelay * pollingBackoffFactor)
      void fetchResultTask(taskId, serial)
    }, pollDelay)
  }

  async function fetchTasks() {
    tasksLoading = true
    const [data, err] = await httpy.get<AIEditTaskResult[]>('/api/my-ai-edits', {
      limit: '10',
      page_id: pageId ? String(pageId) : '',
      page_title: title,
    })
    if (err) {
      showToast(err.message || 'AI 편집 목록을 불러오지 못했습니다.')
      tasks = []
    } else {
      tasks = data ?? []
      if (initialLoad && resultTasks.length > 0) {
        selectedTaskId = String(resultTasks[0].id)
        initialLoad = false
      }
    }
    tasksLoading = false
  }

  function applyResultToWikiEditor() {
    if (!resultOutput || !replaceWikiEditorContent(resultOutput)) {
      showToast('반영할 AI 작성본이 없습니다.')
      return
    }

    if (requestType === 'create') {
      onRequestTypeChange?.('edit')
      showToast('AI 작성본을 위키편집기에 반영했습니다.')
    } else {
      showToast('AI 작성본을 위키편집기에 반영했습니다.')
    }
  }

</script>

{#if resultTasks.length > 0}
  <section class="flex min-h-0 flex-col gap-2 border bg-a-gray-100 p-4">
    <div class="flex items-center gap-2">
      <span class="shrink-0 text-sm">AI 작성본</span>
      <ZSelect bind:value={selectedTaskId} items={taskSelectItems} class="flex-1" placeholder="-- AI 작성본 목록 --">
        {#snippet item(task: TaskSelectItem)}
          <div class="grid min-w-0 flex-1 grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-3">
            <span class="truncate">{task.label}</span>
            <span class="text-right text-xs tabular-nums opacity-50">{task.secondaryLabel ?? ''}</span>
            <span class="text-right text-xs tabular-nums opacity-50">{task.dateLabel ?? ''}</span>
          </div>
        {/snippet}
      </ZSelect>
      {#if tasksLoading}
        <ZSpinner size="0.875rem" />
      {/if}
      <CButton
        type="button"
        variant="ghost"
        size="icon-sm"
        title={resultsExpanded ? '접기' : '펼치기'}
        aria-label={resultsExpanded ? 'AI 작성본 접기' : 'AI 작성본 펼치기'}
        onclick={() => (resultsExpanded = !resultsExpanded)}
      >
        <ZIcon path={resultsExpanded ? mdiChevronUp : mdiChevronDown} />
      </CButton>
    </div>
    {#if selectedTaskId && resultsExpanded}
      <div class="relative">
        <ZTextarea
          id="ai-edit-task-result-textarea"
          class="border-border p-1 min-h-40 bg-a-gray-50 font-mono text-xs leading-relaxed"
          modelValue={resultOutput}
          placeholder="AI 작성본이 여기에 표시됩니다."
          maxHeight={500}
          onUpdateModelValue={(value) => (resultOutput = value)}
        />
        <div class="absolute right-5 top-1 z-10">
          <CButton type="button" variant="subtle" size="small" disabled={!resultOutput} onclick={applyResultToWikiEditor}>
            <ZIcon path={mdiSend} />
            Apply
          </CButton>
        </div>
      </div>
    {/if}
  </section>
{/if}
