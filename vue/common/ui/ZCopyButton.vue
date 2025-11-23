<script setup lang="ts">
import { ref } from 'vue'
import { useClipboard } from '@vueuse/core'

const props = defineProps<{
  text: string
}>()

const copied = ref(false)

const { copy } = useClipboard()

function handleCopy() {
  if (!props.text) return
  copy(props.text).then(() => {
    copied.value = true
    setTimeout(() => copied.value = false, 3000)
  })
}
</script>

<template>
  <button @click="handleCopy"
    class="border border-gray-300 bg-white text-gray-800 text-sm font-medium px-2 py-1 rounded-md hover:bg-gray-50 transition">
    {{ copied ? '✔ Copied!' : '📋 Copy' }}
  </button>
</template>
