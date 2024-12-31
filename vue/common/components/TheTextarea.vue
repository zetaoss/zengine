<script setup lang="ts">
import { ref, watch } from 'vue'

const emit = defineEmits(['update:modelValue'])
const props = defineProps({
  modelValue: { type: String, required: true },
})
const val = ref('')

function handleInput(event: Event) {
  emit('update:modelValue', val.value)
  const target = event.target as HTMLTextAreaElement
  const parent = target.parentNode as HTMLElement
  parent.dataset.replicatedValue = target.value
}

function updateVal(): void {
  val.value = props.modelValue
}

watch(() => props.modelValue, updateVal)
updateVal()
</script>

<template>
  <div class="grow-wrap">
    <textarea v-model="val" aria-label="comment"
      class="resize-none overflow-hidden bg-transparent placeholder-gray-400 hover:placeholder-gray-200 dark:placeholder-gray-500 dark:focus:placeholder-gray-600"
      placeholder="댓글을 남겨보세요" required @input="handleInput" />
  </div>
</template>

<style lang="scss" scoped>
.grow-wrap {
  @apply grid;

  &::after {
    @apply whitespace-pre-wrap invisible;
    content: attr(data-replicated-value) " ";
  }

  &>textarea,
  &::after {
    @apply p-2;
    grid-area: 1 / 1 / 2 / 2;
  }
}
</style>
