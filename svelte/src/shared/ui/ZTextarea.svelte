<script lang="ts">
  import { afterUpdate, onMount } from 'svelte'
  import { createEventDispatcher } from 'svelte'

  export let modelValue = ''
  export let placeholder: string | undefined
  export let id: string
  export let maxHeight: number | undefined

  const dispatch = createEventDispatcher<{ 'update:modelValue': string }>()

  let textareaEl: HTMLTextAreaElement | null = null

  function adjustHeight() {
    if (!textareaEl) return
    const max = maxHeight ?? 200

    textareaEl.style.height = 'auto'
    const scrollHeight = textareaEl.scrollHeight
    const newHeight = Math.min(scrollHeight, max)

    textareaEl.style.height = `${newHeight}px`
    textareaEl.style.overflowY = scrollHeight > max ? 'auto' : 'hidden'
  }

  function handleInput(event: Event) {
    const target = event.target as HTMLTextAreaElement
    dispatch('update:modelValue', target.value)
    queueMicrotask(adjustHeight)
  }

  onMount(adjustHeight)
  afterUpdate(adjustHeight)
</script>

<textarea
  bind:this={textareaEl}
  class="w-full resize-none overflow-y-hidden bg-inherit p-2 outline-none"
  {id}
  value={modelValue}
  {placeholder}
  on:input={handleInput}
></textarea>
