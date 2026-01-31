<script setup lang="ts">
import httpy from '@common/utils/httpy'
import { onMounted, ref } from 'vue'

interface Row {
  title: string
}

interface Data {
  query?: {
    recentchanges: Row[]
  }
}

const rows = ref<Row[]>([])

const load = async () => {
  const [data, err] = await httpy.get<Data>('/w/api.php', {
    format: 'json',
    action: 'query',
    list: 'recentchanges',
    rcprop: 'title',
    rcnamespace: 0,
    rctype: 'new',
    rcshow: '!bot|!anon',
    rclimit: 25,
  })
  if (err) {
    console.error('recentchanges', err)
    return
  }
  rows.value = data?.query?.recentchanges ?? []
}

onMounted(load)
</script>

<template>
  <ul class="py-2 pl-5">
    <li v-for="r in rows" :key="r.title">
      <a :href="`/wiki/${r.title}`">{{ r.title }}</a>
    </li>
  </ul>
</template>
