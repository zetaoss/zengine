<!-- @/components/MapsMap.vue -->
<template>
  <div :style="hostStyle">
    <div ref="mapEl" class="w-full h-full"></div>
  </div>
</template>

<script setup lang="ts">
import L from 'leaflet'
import { computed, onBeforeUnmount, onMounted, ref, useAttrs } from 'vue'

type MapData = {
  width?: string | number
  height?: string | number
  centre?: { lat: number; lon: number }
  zoom?: number | false
  defzoom?: number
  minzoom?: number | false
  maxzoom?: number | false
  scrollwheelzoom?: boolean
}

const attrs = useAttrs()
const mapEl = ref<HTMLDivElement | null>(null)

const raw = computed(() => {
  const v = attrs['data-mw-maps-mapdata']
  return typeof v === 'string' ? v : ''
})

const mapdata = computed<MapData | null>(() => {
  if (!raw.value) return null
  try {
    return JSON.parse(raw.value) as MapData
  } catch {
    return null
  }
})

const hostStyle = computed(() => {
  const d = mapdata.value
  const width = d?.width ?? 'auto'
  const height = d?.height ?? '350px'
  return {
    width: typeof width === 'number' ? `${width}px` : String(width),
    height: typeof height === 'number' ? `${height}px` : String(height),
    overflow: 'hidden',
  } as const
})

let map: L.Map | null = null

onMounted(() => {
  const d = mapdata.value
  if (!d?.centre || !mapEl.value) return

  const zoom = (typeof d.zoom === 'number' ? d.zoom : d.defzoom) ?? 14

  map = L.map(mapEl.value, {
    scrollWheelZoom: d.scrollwheelzoom ?? true,
  }).setView([d.centre.lat, d.centre.lon], zoom)

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(map)

  queueMicrotask(() => map?.invalidateSize())
})

onBeforeUnmount(() => {
  map?.remove()
  map = null
})
</script>
