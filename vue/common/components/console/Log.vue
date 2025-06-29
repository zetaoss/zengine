<!-- eslint-disable vue/multi-word-component-names -->
<script setup lang="ts">
import { defineComponent, h } from 'vue'

const props = defineProps<{
  level: string
  args: unknown[]
}>()

const levelEmoji = {
  log: '',
  error: '❌',
  warn: '⚠️',
  info: 'ℹ️',
  debug: '🐞',
  trace: '🔍'
} as const

const emoji = levelEmoji[props.level as keyof typeof levelEmoji] ?? ''

const RenderArg = defineComponent({
  props: {
    value: { type: null as unknown as () => unknown, required: true }
  },
  setup(props) {
    const seen = new WeakSet<object>()

    const render = (val: unknown): ReturnType<typeof h> => {
      if (val === null) return h('span', { class: 'null' }, 'null')
      if (val === undefined) return h('span', { class: 'undefined' }, 'undefined')
      if (typeof val === 'string') return h('span', { class: 'string' }, `"${val}"`)
      if (typeof val === 'number') return h('span', { class: 'number' }, val.toString())
      if (typeof val === 'boolean') return h('span', { class: 'boolean' }, val.toString())

      if (typeof val === 'object') {
        if (seen.has(val as object)) return h('span', { class: 'circular' }, '[Circular]')
        seen.add(val as object)

        const isArray = Array.isArray(val)
        const entries = isArray
          ? (val as unknown[])
          : Object.entries(val as Record<string, unknown>)

        return h('details', { class: 'detail', open: true }, [
          h(
            'ul',
            { class: 'pl-4' },
            entries.map((entry, i) => {
              const key = isArray ? `[${i}]` : (entry as [string, unknown])[0]
              const value = isArray ? entry : (entry as [string, unknown])[1]

              return h('li', {}, [
                h('span', { class: isArray ? 'graykey' : 'bluekey' }, `${key}: `),
                render(value)
              ])
            })
          )
        ])
      }

      return h('span', {}, String(val))
    }

    return () => render(props.value)
  }
})
</script>

<template>
  <div :class="['log-entry', level]">
    <div class="flex gap-2 items-start">
      <span class="w-5 text-center">
        <span v-if="emoji">{{ emoji }}</span>
        <span v-else class="emoji-placeholder">&nbsp;</span>
      </span>
      <RenderArg v-for="(arg, i) in args" :key="i" :value="arg" />
    </div>
  </div>
</template>

<style scoped>
.log-entry {
  @apply my-1 py-1 px-2 rounded border bg-white dark:bg-zinc-900;
}

.detail>summary {
  display: none;
}

.emoji-placeholder {
  display: inline-block;
  width: 1em;
  opacity: 0;
}
</style>
