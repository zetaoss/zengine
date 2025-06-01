<script setup lang="ts">
import { ref } from 'vue'

import TheModal from '@common/ui/Modal.vue'
import http from '@/utils/http'
import titleExist from '@/utils/mediawiki'

const emit = defineEmits(['close'])

defineProps({
  show: { type: Boolean, required: true },
})

const title = ref('')
const input = ref(null)
const state = ref(0)

let timeoutID: ReturnType<typeof setTimeout>

async function ok() {
  await http.post('/api/write-request', {
    title: title.value,
  })
}

function cancel() {
  emit('close')
}

async function checkExists() {
  state.value = await titleExist(title.value) ? 2 : 3
}

function onInput(event: Event) {
  clearTimeout(timeoutID)
  timeoutID = setTimeout(() => {
    const target = event.target as HTMLInputElement
    title.value = target.value.trim()
    if (title.value.length < 1) {
      state.value = 0
      return
    }
    state.value = 1
    checkExists()
  }, 500)
}
</script>

<template>
  <TheModal :show="show" :ok-disabled="state != 2" @ok="ok" @cancel="cancel">
    <div class="block w-full">
      <h5>새 작성 요청 등록하기</h5>
      <input ref="input" aria-label="title" type="text" class="w-full border rounded p-1 px-2 my-2" placeholder="제목 입력"
        @input="onInput">
      <div v-if="state == 0" class="w-full rounded p-1 px-2 bg-gray-700 text-white">
        요청할 문서명을 입력해주세요.
      </div>
      <div v-else-if="state == 1" class="w-full rounded p-1 px-2 bg-orange-300 text-white">
        '{{ title }}' 문서가 있는지 확인하고 있습니다.
      </div>
      <div v-else-if="state == 2" class="w-full rounded p-1 px-2 bg-green-700 text-white">
        '{{ title }}' 문서를 요청하려면 OK 버튼을 눌러주세요.
      </div>
      <div v-else class="w-full rounded p-1 px-2 bg-orange-700 text-white">
        '{{ title }}' 문서는 이미 있으므로 등록할 수 없습니다.
      </div>
    </div>
  </TheModal>
</template>
