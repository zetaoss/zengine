<script setup lang="ts">
import { ref } from 'vue'

import { CalendarHeatmap } from 'vue3-calendar-heatmap'

const values = ref([{ date: '2018-9-22', count: 6 }])
const endDate = ref('2018-10-10')
</script>
<template>
  UserCal
  <CalendarHeatmap :values="values" :end-date="endDate" :dark-mode="true" />
</template>
