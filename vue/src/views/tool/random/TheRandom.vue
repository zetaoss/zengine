<template>
  <div class="px-4 py-6 max-w-2xl mx-auto space-y-6 text-gray-800 dark:text-gray-100">
    <h2 class="text-3xl font-bold text-center dark:text-white">랜덤 문자열 생성기</h2>

    <!-- Preset Buttons -->
    <div class="flex justify-center flex-wrap gap-2">
      <button v-for="p in presets" :key="p.key" @click="selectPreset(p.key)"
        class="px-4 py-1.5 rounded border transition"
        :class="preset === p.key
          ? 'bg-blue-600 text-white border-blue-600'
          : 'bg-gray-100 dark:bg-gray-800 border-gray-300 dark:border-gray-600 hover:bg-gray-200 dark:hover:bg-gray-700'">
        {{ p.label }}
      </button>
    </div>

    <!-- Length Slider -->
    <div>
      <label class="block font-medium mb-1">문자열 길이: {{ length }}</label>
      <input type="range" min="1" max="64" v-model="length" :disabled="preset === 'uuid'" class="w-full"
        @change="generate" />
    </div>

    <!-- Charset Buttons -->
    <div class="space-y-2">
      <div class="font-medium">문자 집합</div>
      <div class="flex flex-wrap gap-2">
        <button v-for="btn in charsetButtons" :key="btn.key" @click="toggleCharset(btn.key)"
          :class="charBtnClass(isCharsetEnabled(btn.key), isCharsetReadonly())" :aria-disabled="isCharsetReadonly()"
          :title="isCharsetReadonly() ? '이 프리셋에선 수정 불가' : '클릭하여 토글'">
          {{ btn.label }}
        </button>
      </div>
    </div>

    <!-- Generate Button -->
    <button @click="generate" class="w-full py-2 mt-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition">
      새 문자열 생성
    </button>

    <!-- Output History -->
    <div class="space-y-2 pt-4">
      <div v-for="(item, idx) in history" :key="idx"
        class="flex items-center justify-between p-3 rounded font-mono break-all transition-all" :class="[
          idx === 0
            ? 'border-l-4 border-blue-500 bg-blue-50 dark:bg-blue-900/40'
            : 'bg-gray-50 dark:bg-gray-800',
        ]">
        <span class="mr-2 flex-1 break-words">{{ item }}</span>
        <button @click="copyToClipboard(item, idx)"
          class="text-sm px-2 py-1 rounded border border-gray-300 dark:border-gray-600 hover:bg-gray-200 dark:hover:bg-gray-700 transition min-w-[80px] text-center">
          {{ copiedIndex === idx ? '✔ Copied' : '📋 Copy' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

type PresetKey = 'uuid' | 'hex' | 'k8s' | 'alphanum' | 'custom'
type CharsetKey = 'numbers' | 'a-f' | 'a-z' | 'A-Z' | 'dash' | 'symbols'

const copiedIndex = ref<number | null>(null)

const presets = [
  { key: 'uuid', label: 'UUID' },
  { key: 'hex', label: 'Hex' },
  { key: 'k8s', label: 'K8s' },
  { key: 'alphanum', label: 'Alphanum' },
  { key: 'custom', label: 'Custom' },
] as const

const charsetButtons = [
  { key: 'numbers', label: '0-9' },
  { key: 'a-f', label: 'a-f' },
  { key: 'a-z', label: 'a-z' },
  { key: 'A-Z', label: 'A-Z' },
  { key: 'symbols', label: '!@#' },
  { key: 'dash', label: '-' },
] as const

const presetCharsets: Record<PresetKey, CharsetKey[]> = {
  uuid: ['numbers', 'a-f', 'dash'],
  hex: ['numbers', 'a-f'],
  k8s: ['numbers', 'a-z'],
  alphanum: ['numbers', 'a-z', 'A-Z'],
  custom: ['numbers', 'a-z', 'A-Z', 'symbols'],
}

const preset = ref<PresetKey>('uuid')
const length = ref(36)
const enabledCharsets = ref<Set<CharsetKey>>(new Set())
const history = ref<string[]>([])

function charBtnClass(enabled: boolean, readonly: boolean) {
  const base =
    'px-3 py-1 rounded border text-sm transition font-mono min-w-[3.5rem] text-center'

  const active = enabled
    ? 'bg-blue-600 text-white border-blue-600'
    : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 border-gray-300 dark:border-gray-600'

  const disabled = readonly ? 'opacity-60 cursor-not-allowed' : 'hover:brightness-105'

  return [base, active, disabled]
}

function isCharsetEnabled(key: CharsetKey): boolean {
  return enabledCharsets.value.has(key)
}

function isCharsetReadonly(): boolean {
  return preset.value !== 'custom'
}

function toggleCharset(key: CharsetKey) {
  if (preset.value !== 'custom') return
  if (enabledCharsets.value.has(key)) {
    enabledCharsets.value.delete(key)
  } else {
    enabledCharsets.value.add(key)
  }
}

function selectPreset(p: PresetKey) {
  const previous = preset.value
  preset.value = p
  enabledCharsets.value = new Set(presetCharsets[p])

  const fromUuid = previous === 'uuid'
  const toUuid = p === 'uuid'

  // UUID → 다른 프리셋일 때만 길이 초기화
  if (fromUuid || toUuid) {
    switch (p) {
      case 'uuid': length.value = 36; break
      case 'hex': length.value = 16; break
      case 'k8s': length.value = 8; break
      case 'alphanum': length.value = 12; break
      case 'custom': length.value = 16; break
    }
  }

  generate()
}

function generate() {
  const charMap: Record<CharsetKey, string> = {
    numbers: '0123456789',
    'a-f': 'abcdef',
    'a-z': 'abcdefghijklmnopqrstuvwxyz',
    'A-Z': 'ABCDEFGHIJKLMNOPQRSTUVWXYZ',
    symbols: '!@#$%^&*()_+~`|}{[]\\:;?><,./-=',
    dash: '-',
  }

  if (preset.value === 'uuid') {
    history.value.unshift(crypto.randomUUID())
  } else {
    let chars = ''
    for (const key of enabledCharsets.value) {
      chars += charMap[key]
    }
    if (!chars) {
      history.value.unshift('⚠️ 문자 집합을 선택하세요.')
    } else {
      let result = ''
      for (let i = 0; i < length.value; i++) {
        result += chars.charAt(Math.floor(Math.random() * chars.length))
      }
      history.value.unshift(result)
    }
  }

  history.value = history.value.slice(0, 8)
}

function copyToClipboard(text: string, index: number) {
  navigator.clipboard.writeText(text)
  copiedIndex.value = index
  setTimeout(() => {
    if (copiedIndex.value === index) copiedIndex.value = null
  }, 3000)
}

onMounted(() => {
  selectPreset(preset.value)
})
</script>
