<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'

import ZModal from '@common/ui/ZModal.vue'
import CProgressBar from '@common/components/CProgressBar.vue'
import http from '@/utils/http'
import titleExist from '@/utils/mediawiki'

const emit = defineEmits<{
  (e: 'close'): void
}>()

const props = defineProps<{ show: boolean }>()

type State = 'idle' | 'checking' | 'available' | 'exists'

const title = ref('')
const state = ref<State>('idle')
const isSubmitting = ref(false)

const input = ref<HTMLInputElement | null>(null)
const trimmedTitle = computed(() => title.value.trim())

const canCheck = computed(() => trimmedTitle.value.length > 0 && state.value !== 'checking')
const canSubmit = computed(() => state.value === 'available' && !isSubmitting.value)

function reset() {
  title.value = ''
  state.value = 'idle'
  isSubmitting.value = false
}

async function check() {
  if (!trimmedTitle.value) return

  state.value = 'checking'
  try {
    const exists = await titleExist(trimmedTitle.value)
    state.value = exists ? 'exists' : 'available'
  } catch {
    state.value = 'idle'
  }
}

function onInput(e: Event) {
  const val = (e.target as HTMLInputElement).value
  title.value = val
  if (state.value !== 'idle' && state.value !== 'checking') {
    state.value = 'idle'
  }
}

async function ok() {
  if (!canSubmit.value) return

  isSubmitting.value = true
  try {
    await http.post('/api/write-request', { title: trimmedTitle.value })
    emit('close')
    reset()
  } finally {
    isSubmitting.value = false
  }
}

function cancel() {
  emit('close')
  reset()
}

watch(
  () => props.show,
  (show) => {
    if (show) {
      reset()
      nextTick(() => input.value?.focus())
    }
  }
)
</script>

<template>
  <ZModal :show="show" :ok-disabled="!canSubmit" okColor="primary" @ok="ok" @cancel="cancel">
    <div class="w-full">
      <h5 class="mb-3 font-semibold">새 작성 요청 등록하기</h5>
      <div class="flex items-center gap-2">
        <input ref="input" :value="title" type="text" class="flex-1 border rounded p-1 px-2" placeholder="제목 입력"
          @input="onInput" />
        <button type="button"
          class="relative px-3 py-1 rounded border text-sm disabled:opacity-40 flex items-center gap-1"
          :disabled="!canCheck" @click="check">
          중복확인
        </button>
      </div>
      <div class="mt-2 min-h-[6px]">
        <CProgressBar v-if="state === 'checking'" :indeterminate="true" class="h-1" />
        <div v-else-if="state === 'available'" class="h-1 w-full bg-green-500 rounded" />
        <div v-else-if="state === 'exists'" class="h-1 w-full bg-red-500 rounded" />
      </div>
      <div v-if="state === 'exists'" class="mt-2 text-sm text-red-500">
        '{{ trimmedTitle }}' 문서는 이미 있습니다.
      </div>
    </div>
  </ZModal>
</template>
