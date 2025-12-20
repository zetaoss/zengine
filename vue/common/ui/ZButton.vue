<!-- @common/ui/ZButton.vue -->
<script setup lang="ts">
import { type Component,computed, ref, useAttrs } from 'vue'

export type Color = 'default' | 'danger' | 'ghost' | 'primary'

const props = defineProps<{
  as?: string | Component
  disabled?: boolean
  color?: Color
  size?: 'small' | 'medium'
  cooldown?: number
}>()

const emit = defineEmits<{
  (e: 'click', event: MouseEvent): void
}>()

const attrs = useAttrs()

const tag = computed(() => props.as || 'button')

const isCooling = ref(false)

const sizeClasses: Record<string, string> = {
  medium: 'p-2 text-sm',
  small: 'p-0.5 text-[11px]',
}

const base = 'text-[var(--z-text)] inline-flex items-center justify-center rounded transition ring-1 hover:no-underline leading-none'

const colorClasses: Record<string, string> = {
  default: 'bg-[var(--z-btn-bg)] ring-[var(--z-btn-hover)] hover:bg-[var(--z-btn-hover)]',
  danger: 'bg-[var(--z-danger-bg)] ring-[var(--z-danger-hover)] hover:bg-[var(--z-danger-hover)]',
  ghost: 'bg-transparent ring-transparent hover:bg-[var(--z-btn-hover)]',
  primary: 'bg-[var(--z-primary-bg)] ring-[var(--z-primary-hover)] hover:bg-[var(--z-primary-hover)]',
}

const disabledClasses = 'opacity-50 brightness-[.9] cursor-not-allowed pointer-events-none text-[var(--z-btn-text-disabled)]'

const isDisabled = computed(() => props.disabled || isCooling.value)

const classes = computed(() => {
  const color = props.color || 'default'
  const size = props.size || 'medium'

  return [
    base,
    sizeClasses[size],
    colorClasses[color],
    isDisabled.value && disabledClasses,
  ]
    .filter(Boolean)
    .join(' ')
})

function handleClick(event: MouseEvent) {
  if (isDisabled.value) {
    event.preventDefault()
    event.stopImmediatePropagation()
    return
  }

  emit('click', event)

  const t = props.cooldown ?? 500
  if (t > 0) {
    isCooling.value = true
    setTimeout(() => {
      isCooling.value = false
    }, t)
  }
}
</script>

<template>
  <component :is="tag" v-bind="attrs" :class="classes" :disabled="tag === 'button' ? isDisabled : undefined"
    @click="handleClick">
    <slot />
  </component>
</template>
