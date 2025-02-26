<script setup lang="ts">
import ConsoleArg from './ConsoleArg.vue'

defineProps<{
  typ: string
  entries: [string | number, unknown][]
  inEntry: boolean
  depth: number
}>();
</script>

<template>
  <span class="italic">
    <template v-if="typ === 'Array'">
      <template v-if="depth == 0 || (minify == 0 && !expanded)">
        <span>({{ entries.length }})&nbsp;</span>
        <span>[</span>
        <span v-for="[index, value] in entries.slice(0, 5)" :key="index">
          <span v-if="index != 0">, </span>
          <ConsoleArg :arg="value" :depth="depth + 1" :minify="2" :expandable="false" />
        </span>
        <span v-if="entries.length > 5">, …</span>
        <span>]</span>
      </template>
      <template v-else>
        <span>Array({{ entries.length }})</span>
      </template>
    </template>
    <template v-else-if="typ === 'Object'">
      <template v-if="minify == 0">
        <span v-if="name != 'Object'">{{ name }}&nbsp;</span>
        <span>{</span>
        <span v-for="([key, value], index) in entries.slice(0, 5)" :key="index">
          <span v-if="index != 0">, </span>
          <span class="text-gray-500">{{ key }}</span>
          <span>:&nbsp;</span>
          <ConsoleArg :arg="value" :depth="depth + 1" :minify="2" :expandable="false" />
        </span>
        <span v-if="entries.length > 5">, …</span>
        <span>}</span>
      </template>
      <template v-else>
        <span>{{ name }}</span>
      </template>
    </template>
    <template v-else>
      <span>( {{ typ }} )</span>
      <span>{{ entries }}</span>
    </template>
  </span>
</template>
