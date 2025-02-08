<script setup lang="ts">
import { ref, watchEffect } from 'vue'
import ArgObjectArray from './ArgObjectArray.vue'
import ConsoleArg from './ConsoleArg.vue'

const props = defineProps<{
  arg: unknown
  depth: number
  summary: boolean
}>();

const kind = ref('')
const text = ref('')
const entries = ref<[string, unknown][]>([])
const expand = ref(false)

watchEffect(() => {
  const a = props.arg

  if (typeof a === 'object') {
    if (a === null) {
      kind.value = 'null'
      text.value = 'null'
      return
    }
    if (Array.isArray(a)) {
      kind.value = 'array'
      return
    }
    const name = a.constructor.name || (Symbol.toStringTag in a ? 'console' : '');
    if (name !== 'Object') {
      kind.value = 'global'
      text.value = name + ' global{}'
      return
    }
    kind.value = 'plain'
    text.value = '{}'
    entries.value = Object.entries(a)
  }
})
</script>

<template>
  <template v-if="kind == 'array'">
    <ArgObjectArray :arg="arg" :depth="depth" :expand="false" :summary="summary" />
  </template>
  <template v-else-if="kind === 'plain'">
    <span class="bg-blue-500" @click="expand = !expand">
      {{ depth }}
      <template v-if="expand">
        <template v-for="[key, value] in entries" :key="key">
          <span class="object-key">{{ key }}:</span>
          <span class="object-value">
            <ConsoleArg :arg="value" :depth="depth + 1" :expand="false" :summary="true" />
          </span>
        </template>
      </template>
      <template v-else>
        Array({{ entries.length }})
      </template>
    </span>
  </template>
  <template v-else>
    <span :class="kind">{{ text }}</span>
  </template>
</template>

<style scoped lang="scss">
.null {
  @apply text-neutral-400 dark:text-neutral-500;
}
</style>
