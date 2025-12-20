<script setup lang="ts">
import httpy from '@common/utils/httpy'
import { onMounted,ref } from 'vue'

interface Row {
  id: number
  title: string
  replies_count: number
}

const rows = ref<Row[]>([])

const load = async () => {
  const [data, err] = await httpy.get<Row[]>('/api/posts/recent')
  if (err) {
    console.error('recent posts', err)
    return
  }

  rows.value = data
}

onMounted(load)
</script>

<template>
  <div v-for="r in rows" :key="r.id" class="p-1">
    <a :href="`/forum/${r.id}`">
      {{ r.title }}
      <small v-if="r.replies_count > 0">[{{ r.replies_count }}]</small>
    </a>
  </div>
</template>
