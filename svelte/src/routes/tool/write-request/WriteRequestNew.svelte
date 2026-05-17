<script lang="ts">
  import { mdiCheckCircle, mdiCloseCircle, mdiProgressClock } from '@mdi/js'
  import { onDestroy } from 'svelte'
  import { tick } from 'svelte'
  import { createEventDispatcher } from 'svelte'

  import titleExist from '$lib/utils/mediawiki'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZModal from '$shared/ui/ZModal.svelte'
  import httpy from '$shared/utils/httpy'

  type State = 'idle' | 'checking' | 'available' | 'exists'
  const AUTO_CHECK_DELAY_MS = 700

  export let show = false

  const dispatch = createEventDispatcher<{ close: void }>()

  let title = ''
  let state: State = 'idle'
  let isSubmitting = false
  let inputEl: HTMLInputElement | null = null
  let checkedTitle = ''
  let autoCheckTimer: ReturnType<typeof setTimeout> | undefined
  let checkSeq = 0
  $: trimmedTitle = title.trim()
  $: canCheck = trimmedTitle.length > 0 && state !== 'checking' && checkedTitle !== trimmedTitle
  $: canSubmit = state === 'available' && !isSubmitting
  $: status = (() => {
    if (!trimmedTitle) {
      return {
        icon: undefined,
        label: '대기중',
        text: '',
        className: '',
      }
    }
    if (state === 'checking') {
      return {
        icon: mdiProgressClock,
        label: '확인중',
        text: '중복확인',
        className: 'text-blue-600 dark:text-blue-400',
      }
    }
    if (state === 'available') {
      return {
        icon: mdiCheckCircle,
        label: '확인완료',
        text: '중복확인',
        className: 'text-green-600 dark:text-green-400',
      }
    }
    if (state === 'exists') {
      return {
        icon: mdiCloseCircle,
        label: '중복',
        text: '중복확인',
        className: 'text-red-600 dark:text-red-400',
      }
    }
    return {
      icon: undefined,
      label: '대기중',
      text: '',
      className: '',
    }
  })()

  $: if (show) {
    reset()
    void tick().then(() => {
      inputEl?.focus()
    })
  }

  function reset() {
    clearAutoCheck()
    title = ''
    state = 'idle'
    isSubmitting = false
    checkedTitle = ''
    checkSeq += 1
  }

  function clearAutoCheck() {
    if (!autoCheckTimer) return
    clearTimeout(autoCheckTimer)
    autoCheckTimer = undefined
  }

  function scheduleAutoCheck() {
    clearAutoCheck()

    const titleToCheck = title.trim()
    if (!titleToCheck || state === 'checking' || checkedTitle === titleToCheck) return

    autoCheckTimer = setTimeout(() => {
      void check()
    }, AUTO_CHECK_DELAY_MS)
  }

  async function check() {
    clearAutoCheck()

    const titleToCheck = trimmedTitle
    if (!titleToCheck || checkedTitle === titleToCheck) return

    state = 'checking'
    const seq = ++checkSeq

    try {
      const exists = await titleExist(titleToCheck)
      if (seq !== checkSeq || trimmedTitle !== titleToCheck) return
      checkedTitle = titleToCheck
      state = exists ? 'exists' : 'available'
    } catch {
      if (seq !== checkSeq || trimmedTitle !== titleToCheck) return
      state = 'idle'
    }
  }

  function onInput(event: Event) {
    title = (event.target as HTMLInputElement).value
    checkedTitle = ''
    checkSeq += 1
    state = 'idle'
    scheduleAutoCheck()
  }

  function onInputKeydown(event: KeyboardEvent) {
    if (event.key !== 'Enter' || event.isComposing) return

    event.preventDefault()
    if (canSubmit) {
      void ok()
      return
    }

    if (canCheck) {
      void check()
    }
  }

  async function ok() {
    if (!canSubmit) return

    isSubmitting = true
    const [, err] = await httpy.post('/api/write-request', {
      title: trimmedTitle,
    })

    if (err) {
      console.error(err)
      showToast('등록에 실패했습니다. 잠시 후 다시 시도해주세요.')
      reset()
      return
    }

    dispatch('close')
    showToast('등록 완료')
    reset()
  }

  function cancel() {
    dispatch('close')
    reset()
  }

  onDestroy(clearAutoCheck)
</script>

<ZModal {show} title="새 작성 요청 등록하기" okDisabled={!canSubmit} okColor="primary" on:ok={ok} on:cancel={cancel}>
  <div class="w-full">
    <div class="flex items-center gap-2">
      <input
        bind:this={inputEl}
        value={title}
        type="text"
        class="min-w-0 flex-1 rounded border p-1 px-2"
        placeholder="제목 입력"
        on:input={onInput}
        on:keydown={onInputKeydown}
      />
      <div
        class={`flex w-24 shrink-0 items-center justify-start gap-1 text-sm ${status.className}`}
        aria-label={`중복확인 (${status.label})`}
        aria-live="polite"
      >
        {#if status.text}
          <span>{status.text}</span>
        {/if}
        {#if status.icon}
          <ZIcon path={status.icon} class={state === 'checking' ? 'animate-spin' : ''} />
        {/if}
      </div>
    </div>
    {#if state === 'exists'}
      <div class="mt-2 text-sm text-red-500">'{trimmedTitle}' 문서는 이미 있습니다.</div>
    {/if}
  </div>
</ZModal>
