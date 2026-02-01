// @/utils/maps.ts
import { createApp } from 'vue'

import MapsMap from '@/components/MapsMap.vue'

export function mountMaps() {
  const nodes = document.querySelectorAll<HTMLElement>('.maps-map.maps-leaflet')
  for (const el of nodes) {
    if (el.dataset.mounted === '1') continue
    el.dataset.mounted = '1'
    const raw = el.getAttribute('data-mw-maps-mapdata') as string
    const mapdata = JSON.parse(raw)
    createApp(MapsMap, { mapdata }).mount(el)
  }
}
