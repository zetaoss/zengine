// app.ts
import '@/assets/app.css'
import 'leaflet/dist/leaflet.css'

import CAds from '@common/components/CAds.vue'
import CSiteFooter from '@common/components/CSiteFooter.vue'
import CSiteRemocon from '@common/components/CSiteRemocon.vue'
import CSearch from '@common/components/navbar/CSearch.vue'
import { runVisibleHeightsUpdater } from '@common/utils/runVisibleHeightsUpdater'
import { createApp } from 'vue'

import BinderApex from '@/components/binder/BinderApex.vue'
import MapsMap from '@/components/MapsMap.vue'
import UserMenu from '@/components/navbar/UserMenu.vue'
import PageFooter from '@/components/PageFooter.vue'
import PageMenu from '@/components/PageMenu.vue'
import PageMeta from '@/components/PageMeta.vue'
import { runbox } from '@/components/runbox/runbox'
import TocApex from '@/components/toc/TocApex.vue'

runVisibleHeightsUpdater()

const app = createApp({})

app.component('CAds', CAds)
app.component('CSearch', CSearch)
app.component('CSiteFooter', CSiteFooter)
app.component('CSiteRemocon', CSiteRemocon)

app.component('BinderApex', BinderApex)
app.component('MapsMap', MapsMap)
app.component('PageFooter', PageFooter)
app.component('PageMenu', PageMenu)
app.component('PageMeta', PageMeta)
app.component('TocApex', TocApex)
app.component('UserMenu', UserMenu)

app.mount('#app')

runbox()
