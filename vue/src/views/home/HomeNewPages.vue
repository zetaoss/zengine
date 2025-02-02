<script setup lang="ts">
import { ref, onMounted } from 'vue'
import http from '@/utils/http'

interface Row {
  title: string
}

const rows1 = ref<Row[]>([])
const rows2 = ref<Row[]>([])

async function fetchRows(rcshow: string, rclimit: number): Promise<Row[]> {
  const { data } = await http.get('/w/api.php', {
    params: {
      format: 'json',
      action: 'query',
      list: 'recentchanges',
      rcprop: 'title',
      rcnamespace: 0,
      rctype: 'new',
      rcshow,
      rclimit,
    },
  })
  return data.query.recentchanges
}

onMounted(async () => {
  const [res1, res2] = await Promise.all([
    fetchRows('!bot|!anon', 21),
    fetchRows('!bot|anon', 4)
  ])
  rows1.value = res1
  rows2.value = res2
})
</script>

<template>
  <ul class="py-2 pl-5">
    <li v-for="r in rows1" :key="r.title">
      <a :href="`/wiki/${r.title}`">{{ r.title }}</a>
    </li>
  </ul>
  <hr>
  <ul class="py-2 pl-5">
    <li v-for="r in rows2" :key="r.title">
      <a :href="`/wiki/${r.title}`">{{ r.title }}</a>
    </li>
  </ul>
</template>
