<script setup lang="ts">
import { ref } from 'vue'

import linkify from '@/utils/linkify'

const props = defineProps({
  content: { type: String, required: true },
})
const linkifed = ref('')

function fetchData() {
  linkify(props.content).then((v) => {
    linkifed.value = v
  })
}
fetchData()
</script>
<template>
  <span class="message" v-html="linkifed" />
</template>
