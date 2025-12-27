// app.ts
import '@/assets/app.css'

import CAds from '@common/components/CAds.vue'
import CSiteFooter from '@common/components/CSiteFooter.vue'
import CSiteRemocon from '@common/components/CSiteRemocon.vue'
import CNavbarSearch from '@common/components/navbar/CNavbarSearch.vue'
import { runVisibleHeightsUpdater } from '@common/utils/runVisibleHeightsUpdater'
import { createApp } from 'vue'

import BinderApex from '@/components/binder/BinderApex.vue'
import NavbarUserMenu from '@/components/NavbarUserMenu.vue'
import PageFooter from '@/components/PageFooter.vue'
import PageMenu from '@/components/PageMenu.vue'
import PageMeta from '@/components/PageMeta.vue'
import { runbox } from '@/components/runbox/runbox'
import TocApex from '@/components/toc/TocApex.vue'

runVisibleHeightsUpdater()

const app = createApp({})

app.component('CAds', CAds)
app.component('CSiteFooter', CSiteFooter)
app.component('CSiteRemocon', CSiteRemocon)
app.component('CNavbarSearch', CNavbarSearch)

app.component('NavbarUserMenu', NavbarUserMenu)
app.component('PageMenu', PageMenu)
app.component('PageMeta', PageMeta)
app.component('BinderApex', BinderApex)
app.component('PageFooter', PageFooter)
app.component('TocApex', TocApex)

app.mount('#app')

runbox()

