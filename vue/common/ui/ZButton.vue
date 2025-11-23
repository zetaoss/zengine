<script setup lang="ts">
import { computed, useAttrs, type Component } from "vue"

const props = defineProps<{
  as?: string | Component
  disabled?: boolean
  color?: string
}>()

const emit = defineEmits<{
  (e: "click", event: MouseEvent): void
}>()

const attrs = useAttrs()

const tag = computed(() => props.as || "button")

const base =
  "text-[var(--z-text)] inline-flex items-center justify-center rounded transition ring-1 px-2 py-0.5 hover:no-underline"

const colorClasses: Record<string, string> = {
  base: "bg-[var(--z-btn-bg)] ring-[var(--z-btn-hover)] hover:bg-[var(--z-btn-hover)]",
  danger: "bg-[var(--z-danger-bg)] ring-[var(--z-danger-hover)] hover:bg-[var(--z-danger-hover)]",
  ghost: "bg-transparent ring-transparent hover:bg-[var(--z-btn-hover)]",
  primary: "bg-[var(--z-primary-bg)] ring-[var(--z-primary-hover)] hover:bg-[var(--z-primary-hover)]",
}

const disabledClasses = "opacity-40 cursor-not-allowed pointer-events-none text-[var(--z-btn-text-disabled)]"

const classes = computed(() => {
  const color = props.color || "base"
  return [
    base,
    colorClasses[color],
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
  <component :is="tag" v-bind="attrs" :class="classes" :disabled="tag === 'button' ? props.disabled : undefined"
    @click="handleClick">
    <slot />
  </component>
</template>
