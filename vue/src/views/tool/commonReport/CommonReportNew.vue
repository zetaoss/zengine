<script setup lang="ts">
import { nextTick, ref, watch, computed } from 'vue'
import ZModal from '@common/ui/ZModal.vue'
import http from '@/utils/http'

const emit = defineEmits<{
  (e: 'close'): void
}>()

const props = defineProps<{
  show: boolean
}>()

const names = ref<string[]>(Array(4).fill(''))
const inputs = ref<HTMLInputElement[]>([])
const errorMessage = ref('')

const trimmedNames = computed(() =>
  names.value
    .map((name) => name.trim())
    .filter((name) => name !== '')
)

async function ok() {
  // 초기화
  errorMessage.value = ''

  if (trimmedNames.value.length < 2) {
    errorMessage.value = '비교 대상을 2개 이상 입력해 주세요.'
    return
  }

  await http.post('/api/common-report', {
    names: trimmedNames.value,
  })

  emit('close')
}

function cancel() {
  emit('close')
}

watch(
  () => props.show,
  (show) => {
    if (show) {
      nextTick(() => {
        inputs.value[0]?.focus()
      })
    }
  }
)
</script>

<template>
  <ZModal :show="show" @ok="ok" okColor="primary" :okDisabled="trimmedNames.length < 2" @cancel="cancel">
    <div class="block w-full">
      <div class="text-lg mb-2">새로운 비교 등록하기</div>
      <div v-if="errorMessage" class="text-red-600 text-sm mb-2">
        {{ errorMessage }}
      </div>

      <div v-for="(name, index) in names" :key="index" class="pt-2">
        <input ref="inputs" v-model="names[index]" aria-label="word" type="text"
          class="block w-full rounded border px-2 py-1" />
      </div>
    </div>
  </ZModal>
</template>
