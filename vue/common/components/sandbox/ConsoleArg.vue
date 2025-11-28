<!-- ConsoleArg.vue -->
<script setup lang="ts">
import { computed, ref } from 'vue'

defineOptions({ name: 'ConsoleArg' })

const props = defineProps<{
  value: unknown
}>()

const PREVIEW_COUNT = 5

const Kind = {
  Null: 'null',
  Undefined: 'undefined',
  String: 'String',
  Number: 'Number',
  Boolean: 'Boolean',
  Array: 'Array',
  Object: 'Object',
  Function: 'Function',
  Other: 'Other',
} as const

type Kind = (typeof Kind)[keyof typeof Kind]

function getKind(val: unknown): Kind {
  if (val === null) return Kind.Null
  if (val === undefined) return Kind.Undefined

  const t = typeof val

  if (t === 'string') return Kind.String
  if (t === 'number') return Kind.Number
  if (t === 'boolean') return Kind.Boolean
  if (Array.isArray(val)) return Kind.Array
  if (t === 'function') return Kind.Function
  if (t === 'object') return Kind.Object

  return Kind.Other
}

const kind = computed(() => getKind(props.value))

const isSingle = computed(() =>
  kind.value === Kind.Null ||
  kind.value === Kind.Undefined ||
  kind.value === Kind.String ||
  kind.value === Kind.Number ||
  kind.value === Kind.Boolean ||
  kind.value === Kind.Function
)

const isArray = computed(() => kind.value === Kind.Array)
const isObject = computed(() => kind.value === Kind.Object)

const isOpen = ref(false)

type Entry = [string, unknown]

function getEntries(val: unknown, k: Kind): Entry[] {
  if (k === Kind.Array) {
    return (val as unknown[]).map((v, i) => [String(i), v])
  }

  if (k === Kind.Object) {
    return Object.entries(val as Record<string, unknown>)
  }

  return []
}

const entries = computed(() => getEntries(props.value, kind.value))
const preview = computed(() => entries.value.slice(0, PREVIEW_COUNT))
const hasMore = computed(() => entries.value.length > PREVIEW_COUNT)

function getDisplayName(val: unknown, k: Kind): string {
  if (k === Kind.Array) {
    return `Array(${(val as unknown[]).length})`
  }

  if (k === Kind.Object) {
    const ctorName = (val as { constructor?: { name?: string } })?.constructor?.name
    return ctorName || Kind.Object
  }

  return k
}

function formatValue(v: unknown): string {
  const k = getKind(v)

  switch (k) {
    case Kind.Array:
    case Kind.Object:
      return getDisplayName(v, k)
    case Kind.String:
      return `'${v}'`
    case Kind.Function:
      return String(v).replace('function', 'ƒ')
    default:
      return String(v)
  }
}

function getClassFromKind(k: Kind): string {
  return k.toLowerCase()
}

function handleToggle(e: Event) {
  const el = e.target as HTMLDetailsElement
  isOpen.value = el.open
}
</script>

<template>
  <!-- Single -->
  <span v-if="isSingle">
    <span :class="getClassFromKind(kind)">{{ formatValue(value) }}</span>
  </span>

  <!-- Complex(Array/Object) -->
  <template v-else>
    <details class="inline-block align-top" :open="isOpen" @toggle="handleToggle">
      <summary>
        <!-- Array -->
        <template v-if="isArray">
          <span v-if="isOpen">
            {{ getDisplayName(value, kind) }}
          </span>
          <span v-else>
            ({{ (value as unknown[]).length }})
            <span v-if="preview.length">
              <span>[</span>
              <span v-for="([k, v], idx) in preview" :key="k">
                <span :class="getClassFromKind(getKind(v))">{{ formatValue(v) }}</span>
                <span v-if="idx < preview.length - 1">, </span>
              </span>
              <span v-if="hasMore">, …</span>
              <span>]</span>
            </span>
          </span>
        </template>

        <!-- Object -->
        <template v-else-if="isObject">
          {{ getDisplayName(value, kind) }}
          <span v-if="preview.length">
            <span>{</span>
            <span v-for="([k, v], idx) in preview" :key="k">
              <span class="objkey">{{ k }}</span>: {{ formatValue(v) }}
              <span v-if="idx < preview.length - 1">, </span>
            </span>
            <span v-if="hasMore">, …</span>
            <span>}</span>
          </span>
        </template>
      </summary>

      <template v-if="isOpen">
        <ul class="list-none pl-4">
          <li v-for="[key, val] in entries" :key="key" class="m-0">
            <span :class="isArray ? 'arrkey' : 'objkey'">{{ key }}: </span>
            <ConsoleArg :value="val" />
          </li>
        </ul>
      </template>
    </details>
  </template>
</template>
