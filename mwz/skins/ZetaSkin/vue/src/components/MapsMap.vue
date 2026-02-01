<!-- @/components/MapsMap.vue -->
<template>
  <div ref="mapEl" class="h-full"></div>
</template>

<script setup lang="ts">
import L from 'leaflet'
import { onBeforeUnmount, onMounted, ref } from 'vue'

type LatLon = { lat: number; lon: number }

type Location = LatLon & {
  title?: string
  text?: string
  link?: string
  icon?: string
}

type Line = {
  title?: string
  text?: string
  link?: string
  strokeColor?: string
  strokeOpacity?: string | number
  strokeWeight?: string | number
  pos?: LatLon[]
}

type Polygon = {
  title?: string
  text?: string
  link?: string
  strokeColor?: string
  strokeOpacity?: string | number
  strokeWeight?: string | number
  pos?: LatLon[]
  onlyVisibleOnHover?: boolean
  fillColor?: string
  fillOpacity?: string | number
}

type Circle = {
  title?: string
  text?: string
  link?: string
  strokeColor?: string
  strokeOpacity?: string | number
  strokeWeight?: string | number
  fillColor?: string
  fillOpacity?: string | number
  centre?: { lat: number; lon: number } | false
  radius?: number | string
}

type Rectangle = {
  title?: string
  text?: string
  link?: string
  strokeColor?: string
  strokeOpacity?: string | number
  strokeWeight?: string | number
  fillColor?: string
  fillOpacity?: string | number
  ne?: { lat: number; lon: number } | false
  sw?: { lat: number; lon: number } | false
}

type MapData = {
  centre?: false | Location
  locations?: Location[]
  lines?: Line[]
  polygons?: Polygon[]
  circles?: Circle[]
  rectangles?: Rectangle[]
  zoom?: number | false
  defzoom?: number
  minzoom?: number | false
  maxzoom?: number | false
  scrollwheelzoom?: boolean
  static?: boolean
}

const props = defineProps<{ mapdata: MapData }>()
const mapEl = ref<HTMLDivElement | null>(null)
let map: L.Map | null = null

function asNumber(v: unknown): number | null {
  if (typeof v === 'number' && Number.isFinite(v)) return v
  if (typeof v === 'string') {
    const n = Number(v)
    return Number.isFinite(n) ? n : null
  }
  return null
}

function pickCenter(d: MapData): LatLon | null {
  if (d.centre && typeof d.centre === 'object') return d.centre
  const loc0 = d.locations?.[0]
  if (loc0) return loc0
  const lineP0 = d.lines?.[0]?.pos?.[0]
  if (lineP0) return lineP0
  const polyP0 = d.polygons?.[0]?.pos?.[0]
  if (polyP0) return polyP0

  const circle0 = d.circles?.[0]
  if (circle0?.centre && typeof circle0.centre === 'object') return { lat: circle0.centre.lat, lon: circle0.centre.lon }

  const rect0 = d.rectangles?.[0]
  if (rect0?.ne && rect0?.sw && typeof rect0.ne === 'object' && typeof rect0.sw === 'object') {
    return {
      lat: (rect0.ne.lat + rect0.sw.lat) / 2,
      lon: (rect0.ne.lon + rect0.sw.lon) / 2,
    }
  }

  return null
}

function collectBounds(d: MapData): L.LatLngBounds | null {
  const pts: L.LatLngExpression[] = []

  for (const loc of d.locations ?? []) {
    if (typeof loc?.lat === 'number' && typeof loc?.lon === 'number') pts.push([loc.lat, loc.lon])
  }
  for (const line of d.lines ?? []) {
    for (const p of line.pos ?? []) {
      if (typeof p?.lat === 'number' && typeof p?.lon === 'number') pts.push([p.lat, p.lon])
    }
  }
  for (const poly of d.polygons ?? []) {
    for (const p of poly.pos ?? []) {
      if (typeof p?.lat === 'number' && typeof p?.lon === 'number') pts.push([p.lat, p.lon])
    }
  }
  for (const c of d.circles ?? []) {
    if (c?.centre && typeof c.centre === 'object') pts.push([c.centre.lat, c.centre.lon])
  }
  for (const r of d.rectangles ?? []) {
    const ne = r?.ne
    const sw = r?.sw
    if (ne && sw && typeof ne === 'object' && typeof sw === 'object') {
      pts.push([ne.lat, ne.lon], [sw.lat, sw.lon])
    }
  }

  if (pts.length === 0) return null
  const b = L.latLngBounds(pts)
  return b.isValid() ? b : null
}

onMounted(() => {
  const d = props.mapdata
  if (!mapEl.value) return

  const center = pickCenter(d)
  if (!center) return

  const zoom = (typeof d.zoom === 'number' ? d.zoom : d.defzoom) ?? 14
  const minZoom = typeof d.minzoom === 'number' ? d.minzoom : undefined
  const maxZoom = typeof d.maxzoom === 'number' ? d.maxzoom : 19
  const isStatic = d.static === true

  map = L.map(mapEl.value, {
    scrollWheelZoom: isStatic ? false : (d.scrollwheelzoom ?? true),
    dragging: isStatic ? false : true,
    doubleClickZoom: isStatic ? false : true,
    boxZoom: isStatic ? false : true,
    keyboard: isStatic ? false : true,
    zoomControl: !isStatic,
    minZoom,
    maxZoom,
  }).setView([center.lat, center.lon], zoom)

  L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '© OpenStreetMap',
    maxZoom,
  }).addTo(map)

  // locations → marker
  for (const loc of d.locations ?? []) {
    if (typeof loc.lat !== 'number' || typeof loc.lon !== 'number') continue
    const color = loc.icon || 'blue'
    const icon = L.icon({
      ...L.Icon.Default.prototype.options,
      iconUrl: `//unpkg.com/leaflet-color-markers/img/marker-icon-2x-${color}.png`,
      shadowUrl: '//unpkg.com/leaflet/dist/images/marker-shadow.png',
    })
    const marker = L.marker([loc.lat, loc.lon], { icon })
    marker.addTo(map)
    if (loc.title) marker.bindTooltip(loc.title.trim())
  }

  // lines → polyline
  for (const line of d.lines ?? []) {
    const pts = (line.pos ?? [])
      .filter((p): p is LatLon => typeof p?.lat === 'number' && typeof p?.lon === 'number')
      .map((p) => [p.lat, p.lon] as [number, number])

    if (pts.length < 2) continue

    const weight = asNumber(line.strokeWeight) ?? 3
    const opacity = asNumber(line.strokeOpacity) ?? 1
    const color = (line.strokeColor && String(line.strokeColor)) || '#3388ff'

    const layer = L.polyline(pts, { color, weight, opacity }).addTo(map)

    if (line.title) layer.bindTooltip(String(line.title).trim())
  }

  // polygons → polygon
  for (const poly of d.polygons ?? []) {
    const pts = (poly.pos ?? [])
      .filter((p): p is LatLon => typeof p?.lat === 'number' && typeof p?.lon === 'number')
      .map((p) => [p.lat, p.lon] as [number, number])

    if (pts.length < 3) continue

    const weight = asNumber(poly.strokeWeight) ?? 3
    const opacity = asNumber(poly.strokeOpacity) ?? 1
    const color = (poly.strokeColor && String(poly.strokeColor)) || '#3388ff'
    const fillColor = (poly.fillColor && String(poly.fillColor)) || color
    const fillOpacity = asNumber(poly.fillOpacity) ?? 0.2

    const layer = L.polygon(pts, {
      color,
      weight,
      opacity,
      fillColor,
      fillOpacity,
    }).addTo(map)

    if (poly.title) layer.bindTooltip(String(poly.title).trim())
  }

  // circles → circle
  for (const c of d.circles ?? []) {
    const center = c?.centre && typeof c.centre === 'object' ? c.centre : null
    if (!center) continue

    const radius = asNumber(c.radius) ?? 0
    if (!(radius > 0)) continue

    const weight = asNumber(c.strokeWeight) ?? 3
    const opacity = asNumber(c.strokeOpacity) ?? 1
    const color = (c.strokeColor && String(c.strokeColor)) || '#3388ff'
    const fillColor = (c.fillColor && String(c.fillColor)) || color
    const fillOpacity = asNumber(c.fillOpacity) ?? 0.2

    const layer = L.circle([center.lat, center.lon], {
      radius,
      color,
      weight,
      opacity,
      fillColor,
      fillOpacity,
    }).addTo(map)

    if (c.title) layer.bindTooltip(String(c.title).trim())
  }

  // rectangles → rectangle
  for (const r of d.rectangles ?? []) {
    const ne = r?.ne && typeof r.ne === 'object' ? r.ne : null
    const sw = r?.sw && typeof r.sw === 'object' ? r.sw : null
    if (!ne || !sw) continue

    const bounds: L.LatLngBoundsExpression = [
      [sw.lat, sw.lon],
      [ne.lat, ne.lon],
    ]

    const weight = asNumber(r.strokeWeight) ?? 3
    const opacity = asNumber(r.strokeOpacity) ?? 1
    const color = (r.strokeColor && String(r.strokeColor)) || '#3388ff'
    const fillColor = (r.fillColor && String(r.fillColor)) || color
    const fillOpacity = asNumber(r.fillOpacity) ?? 0.2

    const layer = L.rectangle(bounds, {
      color,
      weight,
      opacity,
      fillColor,
      fillOpacity,
    }).addTo(map)

    if (r.title) layer.bindTooltip(String(r.title).trim())
  }

  const bounds = collectBounds(d)
  if (bounds) {
    map.fitBounds(bounds, { padding: [12, 12], maxZoom: zoom })
  }

  requestAnimationFrame(() => map?.invalidateSize())
  setTimeout(() => map?.invalidateSize(), 100)
})

onBeforeUnmount(() => {
  map?.remove()
  map = null
})
</script>
