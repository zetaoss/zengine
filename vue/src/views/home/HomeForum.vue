<script setup lang="ts">
import { ref, onMounted } from 'vue'
import http from '@/utils/http'

interface Row {
  id: number
  title: string
  replies_count: number
}

const rows = ref<Row[]>([])

async function fetchData() {
  try {
    const { data } = await http.get('/api/posts/recent')
    rows.value = data
  } catch (error) {
    console.error('Failed to fetch posts:', error)
  }
}

onMounted(fetchData)
</script>

<template>
  <div v-for="r in rows" :key="r.id" class="p-1">
    <a :href="`/forum/${r.id}`">
      {{ r.title }}
      <small v-if="r.replies_count > 0">[{{ r.replies_count }}]</small>
    </a>
  </div>
</template>
