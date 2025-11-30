// app.ts
import { createApp } from 'vue'

import CSiteFooter from '@common/components/CSiteFooter.vue'
import CSiteRemocon from '@common/components/CSiteRemocon.vue'
import CNavbarSearch from '@common/components/navbar/CNavbarSearch.vue'

import NavbarUserMenu from '@/components/NavbarUserMenu.vue'
import PageMenu from '@/components/PageMenu.vue'
import PageMeta from '@/components/PageMeta.vue'
import BinderApex from '@/components/binder/BinderApex.vue'
import PageFooter from '@/components/pagefooter/PageFooter.vue'
import TocApex from '@/components/toc/TocApex.vue'

import { runVisibleHeightsUpdater } from '@common/utils/runVisibleHeightsUpdater'
import { runbox } from '@/components/runbox/runbox'

import '@/assets/app.scss'

runVisibleHeightsUpdater()

const app = createApp({})

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
