<script setup lang="ts">
import { nextTick, onMounted, watch } from 'vue'

declare global {
  interface Window {
    adsbygoogle?: unknown[]
  }
}

const props = defineProps<{
  client: string
  slot: string
  style?: string
}>()

function pushAd() {
  try {
    const w = window as Window
    w.adsbygoogle = w.adsbygoogle || []
    w.adsbygoogle.push({})
  } catch {
    // ignore
  }
}

async function render() {
  await nextTick()
  pushAd()
}

onMounted(() => {
  render()
})

watch(
  () => props.slot,
  () => {
    render()
  }
)
</script>

<template>
  <ins class="adsbygoogle" :style="style || 'display:inline-block'" :data-ad-client="client" :data-ad-slot="slot" />
</template>
