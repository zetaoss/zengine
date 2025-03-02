<script setup lang="ts">
import { ref, computed } from 'vue';

const props = defineProps<{ text: string }>();
const expanded = ref(false);
const maxLength = 500;

const isTruncated = computed(() => props.text.length > maxLength);
const displayText = computed(() => (expanded.value || !isTruncated.value) ? props.text : props.text.slice(0, maxLength) + '...');

const copyText = () => {
  navigator.clipboard.writeText(props.text).then(() => {
    alert('Copied to clipboard!');
  });
};
</script>

<template>
  <span class="string break-all">
    "{{ displayText.replace(/"/g, '\\"') }}"
  </span>
  <button v-if="isTruncated" @click="expanded = !expanded" class="ml-2 text-blue-500 underline">
    {{ expanded ? 'See less' : 'See more' }}
  </button>
  <button v-if="isTruncated" @click="copyText" class="ml-2 px-2 py-1 text-white bg-gray-700 rounded">Copy</button>
</template>
