<script setup lang="ts">
import { ref, computed } from "vue";

const props = defineProps<{ text: string; inEntry?: boolean }>();

const expanded = ref(false);
const truncateLimit = props.inEntry ? 255 : 100;
const isTruncated = props.text.length > truncateLimit;
const escapedText = computed(() => JSON.stringify(props.text));

const truncatedText = computed(() => {
  if (!isTruncated || expanded.value) return props.inEntry ? escapedText.value : `'${props.text}'`;
  return props.inEntry
    ? escapedText.value.slice(0, 50)
    : `'${props.text.slice(0, 50)}…${props.text.slice(-49)}'`;
});

const sizeInMB = computed(() => props.text.length / (1024 * 1024));
const showMoreText = computed(() => {
  if (!isTruncated) return "";
  return expanded.value ? "Show less" : `Show more (${sizeInMB.value.toFixed(sizeInMB.value >= 0.1 ? 1 : 2)} MB)`;
});

const copyToClipboard = () => {
  navigator.clipboard.writeText(escapedText.value);
};
</script>

<template>
  <span class="text-sky-500 dark:text-sky-300">
    {{ truncatedText }}
  </span>
  <template v-if="inEntry && isTruncated">
    <button @click="expanded = !expanded" class="text-gray-500 hover:text-gray-700 text-xs underline">
      {{ showMoreText }}
    </button>
    <button @click="copyToClipboard" class="text-gray-500 hover:text-gray-700 text-xs underline">
      Copy
    </button>
  </template>
</template>
