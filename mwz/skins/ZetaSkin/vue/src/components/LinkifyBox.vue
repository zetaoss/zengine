<!-- LinkifyBox.vue  -->
<script setup lang="ts">
import { ref, watch } from 'vue'

import linkify from '@/utils/linkify'

const props = defineProps({
  content: { type: String, required: true },
})
const linkifed = ref('')

watch(
  () => props.content,
  (content) => {
    linkify(content).then((v) => {
      linkifed.value = v
    })
  },
  { immediate: true }
)
</script>
<template>
  <span class="message" v-html="linkifed" />
</template>
