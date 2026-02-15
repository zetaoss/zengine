<script lang="ts">
  import { tick } from 'svelte'

  import { showToast } from '$shared/ui/toast/toast'
  import ZModal from '$shared/ui/ZModal.svelte'
  import httpy from '$shared/utils/httpy'

  export let show = false
  export let onClose: (() => void) | undefined = undefined

  let names = ['', '', '', '']
  let errorMessage = ''
  let inputs: HTMLInputElement[] = []
  $: trimmedNames = names.map((name) => name.trim()).filter((name) => name !== '')

  $: if (show) {
    void tick().then(() => {
      inputs[0]?.focus()
    })
  }

  async function ok() {
    errorMessage = ''

    if (trimmedNames.length < 2) {
      errorMessage = '비교 대상을 2개 이상 입력해 주세요.'
      return
    }

    const [, err] = await httpy.post('/api/common-report', {
      names: trimmedNames,
    })

    if (err) {
      console.error(err)
      errorMessage = '등록에 실패했습니다. 잠시 후 다시 시도해 주세요.'
      return
    }

    showToast('등록 완료')
    onClose?.()
  }

  function cancel() {
    onClose?.()
  }
</script>

<ZModal {show} on:ok={ok} okColor="primary" backdropClosable={false} okDisabled={trimmedNames.length < 2} on:cancel={cancel}>
  <div class="block w-full">
    <div class="mb-2 text-lg">새로운 비교 등록하기</div>
    {#if errorMessage}
      <div class="mb-2 text-sm text-red-600">
        {errorMessage}
      </div>
    {/if}

    {#each names as name, index (index)}
      <div class="pt-2">
        <input
          bind:this={inputs[index]}
          bind:value={names[index]}
          aria-label="word"
          type="text"
          class="block w-full rounded border px-2 py-1"
          placeholder={`비교 대상 ${index + 1}${name ? '' : ''}`}
        />
      </div>
    {/each}
  </div>
</ZModal>
