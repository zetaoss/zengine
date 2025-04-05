<script setup lang="ts">
import { stringify } from './utils';

defineProps<{
  level: string;
  arg: unknown;
}>();

</script>

<template>
  <div class="arg-summary">
    <template v-if="typeof arg === 'string'">
      <span class="string">"{{ arg }}"</span>
    </template>
    <template v-else-if="typeof arg === 'number'">
      <span class="number">{{ arg }}</span>
    </template>
    <template v-else-if="typeof arg === 'boolean'">
      <span class="boolean">{{ arg }}</span>
    </template>
    <template v-else-if="arg === null">
      <span class="null">null</span>
    </template>
    <template v-else-if="arg === undefined">
      <span class="undefined">undefined</span>
    </template>
    <template v-else-if="Array.isArray(arg)">
      <span class="array">[{{ arg.join(', ') }}]</span>
    </template>
    <template v-else-if="typeof arg === 'object'">
      <span class="object">{{ stringify(arg) }}</span>
    </template>
  </div>
</template>
