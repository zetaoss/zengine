<script setup lang="ts">
import { computed } from 'vue';
import TheArg from './TheArg.vue';

const props = defineProps<{
  depth: number;
  entries: [string | number, unknown][]
  seen: Set<unknown>;
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
    <template v-for="(entry, i) in sortedEntries" :key="i">
      <div class="flex-grow-0">
        <div class="overflow-auto break-all" :style="{ paddingLeft: `${0.2 * depth + 0.5}rem` }">
          <span>- {{ entry[0] }}:&nbsp;</span>
          <TheArg :arg="entry[1]" :depth="depth + 1" :seen="seen" :inEntry="true" />
        </div>
      </div>
    </template>
  </span>
</template>
