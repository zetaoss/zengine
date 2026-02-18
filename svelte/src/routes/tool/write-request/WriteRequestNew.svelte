<script lang="ts">
  import { tick } from 'svelte'
  import { createEventDispatcher } from 'svelte'

  import titleExist from '$lib/utils/mediawiki'
  import { showToast } from '$shared/ui/toast/toast'
  import ZModal from '$shared/ui/ZModal.svelte'
  import httpy from '$shared/utils/httpy'

  type State = 'idle' | 'checking' | 'available' | 'exists'

  export let show = false

  const dispatch = createEventDispatcher<{ close: void }>()

  let title = ''
  let state: State = 'idle'
  let isSubmitting = false
  let inputEl: HTMLInputElement | null = null
  $: trimmedTitle = title.trim()
  $: canCheck = trimmedTitle.length > 0 && state !== 'checking'
  $: canSubmit = state === 'available' && !isSubmitting

  $: if (show) {
    reset()
    void tick().then(() => {
      inputEl?.focus()
    })
  }

  function reset() {
    title = ''
    state = 'idle'
    isSubmitting = false
  }

  async function check() {
    if (!trimmedTitle) return

    state = 'checking'
    try {
      const exists = await titleExist(trimmedTitle)
      state = exists ? 'exists' : 'available'
    } catch {
      state = 'idle'
    }
  }

  function onInput(event: Event) {
    title = (event.target as HTMLInputElement).value
    if (state !== 'idle' && state !== 'checking') {
      state = 'idle'
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
</script>

<ZModal {show} okDisabled={!canSubmit} okColor="primary" on:ok={ok} on:cancel={cancel}>
  <div class="w-full">
    <h5 class="mb-3 font-semibold">새 작성 요청 등록하기</h5>
    <div class="flex items-center gap-2">
      <input
        bind:this={inputEl}
        value={title}
        type="text"
        class="flex-1 rounded border p-1 px-2"
        placeholder="제목 입력"
        on:input={onInput}
      />
      <button
        type="button"
        class="relative flex items-center gap-1 rounded border px-3 py-1 text-sm disabled:opacity-40"
        disabled={!canCheck}
        on:click={check}
      >
        중복확인
      </button>
    </div>
    <div class="mt-2 min-h-1.5">
      {#if state === 'checking'}
        <div class="progress-wrap h-1">
          <div class="progress-bar"></div>
        </div>
      {:else if state === 'available'}
        <div class="h-1 w-full rounded bg-green-500"></div>
      {:else if state === 'exists'}
        <div class="h-1 w-full rounded bg-red-500"></div>
      {/if}
    </div>
    {#if state === 'exists'}
      <div class="mt-2 text-sm text-red-500">'{trimmedTitle}' 문서는 이미 있습니다.</div>
    {/if}
  </div>
</ZModal>

<style>
  .progress-wrap {
    overflow: hidden;
    width: 100%;
    background: rgba(5, 114, 206, 0.05);
  }

  .progress-bar {
    width: 100%;
    height: 100%;
    animation: indeterminateAnimation 1s infinite linear;
    background: rgb(5, 114, 206);
    transform-origin: 0% 50%;
  }

  @keyframes indeterminateAnimation {
    0% {
      transform: translateX(0) scaleX(0);
    }

    40% {
      transform: translateX(0) scaleX(0.4);
    }

    100% {
      transform: translateX(100%) scaleX(0.5);
    }
  }
</style>
