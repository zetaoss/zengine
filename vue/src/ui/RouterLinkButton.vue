<!-- RouterLinkButton.vue -->
<script setup lang="ts">
import ZButton, { type Color } from "@common/ui/ZButton.vue"
import { useAttrs } from "vue"
import { type RouteLocationRaw,RouterLink } from "vue-router"

const props = defineProps<{
  to: RouteLocationRaw
  replace?: boolean
  disabled?: boolean
  color?: Color
}>()

const emit = defineEmits<{
  (e: "click", event: MouseEvent): void
}>()

const attrs = useAttrs()

function handleClick(event: MouseEvent, navigate: () => void) {
  if (props.disabled) {
    event.preventDefault()
    event.stopImmediatePropagation()
    return
  }

  emit("click", event)

  if (event.defaultPrevented) return

  navigate()
}
</script>

<template>
  <RouterLink v-slot="{ navigate, href }" :to="props.to" :replace="props.replace" custom>
    <ZButton as="a" v-bind="attrs" :href="href" :color="props.color" :disabled="props.disabled"
      @click="(event: MouseEvent) => handleClick(event, navigate)">
      <slot />
    </ZButton>
  </RouterLink>
</template>
