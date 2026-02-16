<script lang="ts">
  import { onMount } from 'svelte'

  export let modelValue = ''
  export let placeholder: string | undefined
  export let id: string
  export let maxHeight: number | undefined
  export let onUpdateModelValue: ((value: string) => void) | undefined = undefined

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
    onUpdateModelValue?.(target.value)
    queueMicrotask(adjustHeight)
  }

  function scheduleAdjust(deps: { modelValue: string; maxHeight: number | undefined }) {
    // Keep this reactive dependency explicit for auto-height recalculation.
    if (!deps) return
    queueMicrotask(adjustHeight)
  }

  onMount(adjustHeight)
  $: scheduleAdjust({ modelValue, maxHeight })
</script>

<textarea
  bind:this={textareaEl}
  class="w-full resize-none overflow-y-hidden bg-inherit p-2 outline-none"
  {id}
  value={modelValue}
  {placeholder}
  on:input={handleInput}
></textarea>
