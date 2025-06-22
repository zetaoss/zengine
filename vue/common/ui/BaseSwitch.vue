<!-- BaseSwitch.vue -->
<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  modelValue: boolean
}>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const isOn = computed(() => props.modelValue)
function toggle() {
  emit('update:modelValue', !isOn.value)
}
</script>

<template>
  <button :class="[
    'w-10 h-6 flex items-center rounded-full transition-colors duration-200',
    isOn ? 'bg-blue-500' : 'bg-gray-300',
  ]" @click="toggle">
    <div :class="[
      'w-4 h-4 bg-white rounded-full shadow transform transition-transform duration-200',
      isOn ? 'translate-x-4' : 'translate-x-1',
    ]" />
  </button>
</template>
