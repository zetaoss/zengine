<script setup lang="ts">
import { RouterLink } from "vue-router"
import { computed } from "vue"

const props = defineProps<{
  to?: string | object
  as?: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: "click", event: MouseEvent): void
}>()

const tag = computed(() => {
  if (props.to) return RouterLink
  return props.as || "button"
})

const base = "inline-flex items-center justify-center rounded transition text-[var(--btn-text)] bg-[var(--btn-bg)] ring-1 ring-[var(--btn-ring)] px-2 py-0.5 hover:no-underline hover:bg-[var(--btn-bg-hover)]"
const disabledClasses = "opacity-40 cursor-not-allowed pointer-events-none text-[var(--btn-text-disabled)]"

const classes = computed(() => {
  return [
    base,
    props.disabled && disabledClasses,
  ]
    .filter(Boolean)
    .join(" ")
})

function handleClick(event: MouseEvent) {
  if (props.disabled) {
    event.preventDefault()
    event.stopImmediatePropagation()
    return
  }
  emit("click", event)
}
</script>

<template>
  <component :is="tag" :to="tag === RouterLink ? props.to : undefined" :class="classes" :disabled="props.disabled"
    @click="handleClick">
    <slot />
  </component>
</template>
