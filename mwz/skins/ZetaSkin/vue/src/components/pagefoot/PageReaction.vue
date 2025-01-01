<script setup lang="ts">
import { ref } from 'vue'

import TheIcon from '@common/components/TheIcon.vue'
import { mdiEmoticonHappyOutline } from '@mdi/js'
import { vOnClickOutside } from '@vueuse/components'

import http from '@/utils/http'
import getRLCONF from '@/utils/rlconf'

const pageid = getRLCONF().wgArticleId
const show = ref(false)
const emojis = ['👍', '❤️', '😆', '😮', '😢']

const emojiCount = ref({} as Object)
const userEmojis = ref([] as string[])

function openMenu() {
  show.value = true
}

function closeMenu() {
  show.value = false
}

async function fetchData() {
  show.value = false
  const data: any = await http.get(`/api/reactions/page/${pageid}`)
  if (data.data === null) return
  emojiCount.value = data.data.emojiCount
  userEmojis.value = data.data.userEmojis
}

function toggleReaction(emoji: string) {
  http.post('/api/reactions/page', {
    pageid,
    emoji,
    enable: userEmojis.value.indexOf(emoji) === -1,
  }).then(() => fetchData())
}
fetchData()
</script>
<template>
  <div class="py-3 text-sm">
    <div class="relative" v-on-click-outside="closeMenu">
      <ul class="mb-2 absolute flex list-none border p-0 bottom-0 rounded z-40" v-if="show">
        <li v-for="emoji in emojis" :key="emoji">
          <button type="button" class="mx-1 p-1 px-2 rounded hover:bg-gray-200 dark:hover:bg-gray-700"
            :class="{ 'bg-gray-100 dark:bg-slate-800': userEmojis.indexOf(emoji) !== -1 }" @click="toggleReaction(emoji)">
            {{ emoji }}
          </button>
        </li>
      </ul>
    </div>
    <button type="button"
      class="item border leading-[0] p-[4px] my-[4px] mx-[2px] rounded-full text-gray-500 hover:bg-gray-200 dark:hover:bg-gray-800"
      @click="openMenu">
      <TheIcon :path='mdiEmoticonHappyOutline' :size="16" />
    </button>
    <span v-for="(cnt, emoji) in emojiCount" :key="emoji">
      <button type="button" class="border leading-[0] p-2 py-3 mx-[2px] rounded-full"
        :class="{ 'border-blue-500 dark:border-blue-500 text-blue-500 bg-blue-50 dark:bg-blue-950': userEmojis.indexOf(emoji) !== -1 }"
        @click="toggleReaction(emoji)">
        {{ emoji }} {{ cnt }}
      </button>
    </span>
  </div>
</template>
