<script setup lang="ts">
import { ref } from 'vue'
import http from '@/utils/http'

interface Row {
  title: string
}

const rows1 = ref([] as Row[])
const rows2 = ref([] as Row[])

async function fetchRows(rcshow: string, rclimit: number) {
  const resp: any = await http.get('/w/api.php', {
    params: {
      format: 'json',
      action: 'query',
      list: 'recentchanges',
      rcprop: 'title',
      rcnamespace: '0',
      rctype: 'new',
      rcshow,
      rclimit: rclimit.toString(),
    },
  })
  return resp.data.query.recentchanges
}

async function fetchData() {
  rows1.value = await fetchRows('!bot|!anon', 21)
  rows2.value = await fetchRows('!bot|anon', 4)
}

fetchData()
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
