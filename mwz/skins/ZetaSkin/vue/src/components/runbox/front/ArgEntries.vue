<script setup lang="ts">
import ConsoleArg from './ConsoleArg.vue'

defineProps<{
  typ: string
  entries: [string | number, unknown][]
  minify: number
  contructor: string
}>();
</script>

<template>
  <span class="italic">
    <template v-if="typ === 'Array'">
      <template v-if="minify == 0">
        <span>({{ entries.length }})&nbsp;</span>
        <span>[</span>
        <span v-for="[index, value] in entries.slice(0, 5)" :key="index">
          <span v-if="index != 0">, </span>
          <ConsoleArg :arg="value" :minify="2" :expandable="false" />
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
        <span>{{ contructor }}&nbsp;</span>
        <span>{</span>
        <span v-for="([key, value], index) in entries.slice(0, 5)" :key="index">
          <span v-if="index != 0">, </span>
          <span class="text-gray-500">{{ key }}</span>
          <span>:&nbsp;</span>
          <ConsoleArg :arg="value" :minify="2" :expandable="false" />
        </span>
        <span v-if="entries.length > 5">, …</span>
        <span>}</span>
      </template>
      <template v-else>
        <span>{{ contructor }}</span>
      </template>
    </template>
    <template v-else>
      <span>( {{ typ }} )</span>
      <span>{{ entries }}</span>
    </template>
  </span>
</template>
