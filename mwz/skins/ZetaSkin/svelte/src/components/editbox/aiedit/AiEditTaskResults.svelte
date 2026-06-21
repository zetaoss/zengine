<svelte:options runes={true} />

<script lang="ts">
  import { mdiRefresh } from '@mdi/js'
  import { onDestroy } from 'svelte'

  import CButton from '$shared/ui/CButton.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSelect from '$shared/ui/ZSelect.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'
  import { getAge } from '$shared/utils/time'

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
    onRequestTypeChange,
  }: {
    pageId: number | undefined
    title: string
    requestType: AIEditRequestType
    submittedTaskId: number | undefined
    resetToken: number
    onPollingChange: (polling: boolean) => void
    onRequestTypeChange?: (newType: AIEditRequestType) => void
  } = $props()

  const pollingInitialMs = 1000
  const pollingBackoffFactor = 1.1
  const activeTaskPhases: ReadonlySet<AIEditTaskResult['phase']> = new Set(['Pending', 'Generating', 'Retrying'])

  let tasks = $state<AIEditTaskResult[]>([])
  let selectedTaskId = $state('')
  let resultOutput = $state('')
  let tasksLoading = $state(false)
  let initialLoad = $state(true)
  let handledSubmittedTaskId: number | undefined
  let handledResetToken: number | undefined
  let pollTimer: ReturnType<typeof setTimeout> | undefined
  let pollDelay = pollingInitialMs
  let pollSerial = 0

  let taskSelectItems = $derived(
    tasks.map((task) => {
      const age = getAge(task.created_at)
      return {
        value: String(task.id),
        label: `#${task.id}${age ? ` (${age})` : ''}`,
      }
    }),
  )

  $effect(() => {
    void fetchTasks()
  })

  $effect(() => {
    if (handledResetToken === undefined) {
      handledResetToken = resetToken
      return
    }
    if (resetToken === handledResetToken) return
    handledResetToken = resetToken
    handledSubmittedTaskId = undefined
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
    resultOutput = tasks.find((task) => String(task.id) === selectedTaskId)?.llm_output ?? ''
  })

  $effect(() => {
    const task = tasks.find((item) => String(item.id) === selectedTaskId)
    selectedAiEditResult.set(task ? { content: resultOutput, taskId: task.id } : null)
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
      if (initialLoad && tasks.length > 0) {
        selectedTaskId = String(tasks[0].id)
        initialLoad = false
      }
    }
    tasksLoading = false
  }

  async function applyResultToWikiEditor() {
    if (!resultOutput || !replaceWikiEditorContent(resultOutput)) {
      showToast('반영할 AI 편집본이 없습니다.')
      return
    }

    if (requestType === 'create') {
      const ok = await showConfirm('위키편집기에 내용이 반영되었습니다. 편집 모드로 전환할까요?')
      if (ok) onRequestTypeChange?.('edit')
    } else {
      showToast('AI 편집본을 위키편집기에 반영했습니다.')
    }
  }
</script>

<section class="flex min-h-0 flex-col gap-2">
  <div class="flex items-center gap-2">
    <ZSelect bind:value={selectedTaskId} items={taskSelectItems} class="flex-1" placeholder="-- AI 편집본 목록 --" />
    {#if tasksLoading}
      <ZSpinner size="0.875rem" />
    {/if}
    <CButton type="button" variant="outline" size="small" title="새로고침" onclick={() => void fetchTasks()} disabled={tasksLoading}>
      <ZIcon path={mdiRefresh} />
    </CButton>
  </div>
  {#if selectedTaskId}
    <textarea
      class="z-input min-h-80 w-full resize-y font-mono text-xs"
      bind:value={resultOutput}
      placeholder="AI 편집본이 여기에 표시됩니다."
    ></textarea>
    <div class="flex justify-start">
      <CButton type="button" variant="default" disabled={!resultOutput} onclick={applyResultToWikiEditor}>위키편집기에 반영</CButton>
    </div>
  {/if}
</section>
