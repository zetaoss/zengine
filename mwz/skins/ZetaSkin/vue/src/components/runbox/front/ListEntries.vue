<script setup lang="ts">
import { computed } from 'vue';
import ConsoleArg from './ConsoleArg.vue';

const props = defineProps<{
  typ: string;
  depth: number;
  minify: number;
  expanded: boolean;
  expandable: boolean;
  entries: [string | number, unknown][]
}>();

const sortedEntries = computed(() => {
  return [...props.entries].sort(([a], [b]) => {
    if (typeof a === 'number' && typeof b === 'number') {
      return a - b;
    }
    return String(a).localeCompare(String(b));
  });
});
</script>

<template>
  <span class="bg-green-900 dummy">
    <template v-if="expandable">
      <template v-if="typ === 'Array'">
        <template v-if="minify == 0">
          <div v-for="[index, value] in entries" :key="index" :style="{ paddingLeft: `${0.2 * depth + 0.5}rem` }">
            <div class="align-text-top">
              <ConsoleArg :k="index" :arg="value" :depth="depth + 1" :minify="2" :expandable="true" />
            </div>
          </div>
        </template>
      </template>
      <template v-else-if="typ === 'Object'">
        <template v-if="minify == 0">
          <div v-for="([key, value], index) in sortedEntries" :key="index"
            :style="{ paddingLeft: `${0.2 * depth + 0.5}rem` }">
            <div class="align-text-top">
              <span>{{ key }}:&nbsp;</span>
              <ConsoleArg :arg="value" :depth="depth + 1" :minify="2" :expandable="true" />
            </div>
          </div>
        </template>
        <template v-else>
          <span>ELSE45</span>
        </template>
      </template>
      <template v-else>
        <span>( {{ typ }} )</span>
        <span>{{ entries }}</span>
      </template>
    </template>
  </span>
</template>
