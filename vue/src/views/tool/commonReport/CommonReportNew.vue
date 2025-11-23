<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'

import ZModal from '@common/ui/ZModal.vue'
import http from '@/utils/http'

const emit = defineEmits(['close'])

const props = defineProps({
  show: { type: Boolean, required: true },
})

const names = [ref(''), ref(''), ref(''), ref('')]
const inputs = ref([] as HTMLElement[])

async function ok() {
  const filtered = names.filter((x) => x.value).map((x) => x.value)
  if (names.length < 2) {
    alert('비교 대상을 2개 이상 입력해 주세요')
    return
  }
  await http.post('/api/common-report', {
    names: filtered,
  })
  emit('close')
}

function cancel() {
  emit('close')
}

watch(() => props.show, (newValue, oldValue) => {
  if (newValue && !oldValue) {
    nextTick(() => {
      inputs.value[0].focus()
    })
  }
})
</script>

<template>
  <ZModal :show="show" @ok="ok" @cancel="cancel">
    <div class="block w-full">
      <h5>새로운 비교 등록하기</h5>
      <hr class="border-t">
      <hr class="border-0">
      <hr class="border-0">

      <div v-for="(name, index) in names" :key="index" class="pt-2">
        <input ref="inputs" v-model="name.value" aria-label="word" type="text"
          class="block w-full border rounded p-1 px-2">
      </div>
    </div>
  </ZModal>
</template>
