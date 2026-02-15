<script lang="ts">
  import { onMount } from 'svelte'

  import ZButton from '$shared/ui/ZButton.svelte'

  import type { ForumPostFormValue } from './types'

  export let modelValue: ForumPostFormValue
  export let submitText: string
  export let submitting = false
  export let titleError: string | null = null
  export let bodyError: string | null = null

  export let onSubmit: () => void = () => {}
  export let onCancel: () => void = () => {}
  export let onClearTitleError: () => void = () => {}
  export let onClearBodyError: () => void = () => {}

  let EditorApex: (typeof import('../editor/EditorApex.svelte'))['default'] | null = null

  onMount(async () => {
    const mod = await import('../editor/EditorApex.svelte')
    EditorApex = mod.default
  })

  let title = modelValue.title
  let body = modelValue.body

  $: if (modelValue.title !== title) title = modelValue.title
  $: if (modelValue.body !== body) body = modelValue.body

  $: disabled = !title.trim() || !body.trim() || submitting

  function update(patch: Partial<ForumPostFormValue>) {
    modelValue = { ...modelValue, ...patch }
  }
</script>

<div class="space-y-4">
  <div class="flex items-center">
    <select
      value={modelValue.cat}
      on:change={(e) => update({ cat: (e.target as HTMLSelectElement).value })}
      class="my-3 w-auto rounded border bg-white p-1 text-sm text-gray-900 dark:bg-black dark:text-gray-100"
    >
      <option value="질문">질문</option>
      <option value="잡담">잡담</option>
      <option value="인사">인사</option>
      <option value="기타">기타</option>
    </select>
  </div>

  <div>
    <input
      value={title}
      on:input={(e) => {
        onClearTitleError()
        title = (e.target as HTMLInputElement).value
        update({ title })
      }}
      type="text"
      placeholder="제목을 입력해 주세요"
      class={`block w-full rounded border bg-white px-4 py-2 text-gray-900 outline-0 dark:bg-black dark:text-gray-300 ${
        titleError ? 'border-red-300 dark:border-red-700' : ''
      }`}
    />
    {#if titleError}
      <div class="text-sm text-red-400">{titleError}</div>
    {/if}
  </div>

  <div>
    <div class={bodyError ? 'border border-red-300 dark:border-red-700' : ''}>
      {#if EditorApex}
        <svelte:component
          this={EditorApex}
          modelValue={body}
          onModelValueChange={(value: string) => {
            body = value
            onClearBodyError()
            update({ body })
          }}
        />
      {:else}
        <div class="p-4 text-sm text-gray-500">에디터 로딩 중...</div>
      {/if}
    </div>
    {#if bodyError}
      <div class="text-sm text-red-400">{bodyError}</div>
    {/if}
  </div>

  <div class="my-4 flex justify-center gap-3">
    <ZButton color="primary" {disabled} onclick={onSubmit}>
      {submitText}
    </ZButton>
    <ZButton onclick={onCancel}>취소</ZButton>
  </div>
</div>
