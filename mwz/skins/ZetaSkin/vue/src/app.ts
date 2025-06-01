import { createApp } from 'vue'

import CLayoutFoot from '@common/components/CLayoutFoot.vue'
import CLayoutRemocon from '@common/components/CLayoutRemocon.vue'
import CNavbarSearch from '@common/components/navbar/CNavbarSearch.vue'

import NavbarUserMenu from '@/components/NavbarUserMenu.vue'
import PageMenu from '@/components/PageMenu.vue'
import PageMeta from '@/components/PageMeta.vue'
import RouterLink from '@/components/RouterLink.vue'
import Binder from '@/components/binder/Binder.vue'
import PageFoot from '@/components/pagefoot/PageFoot.vue'
import TocMain from '@/components/toc/TocMain.vue'
import { runbox } from '@/components/runbox/runbox'

import '@/assets/app.scss'

const app = createApp({})

app.component('CLayoutFoot', CLayoutFoot)
app.component('CLayoutRemocon', CLayoutRemocon)
app.component('CNavbarSearch', CNavbarSearch)

app.component('NavbarUserMenu', NavbarUserMenu)
app.component('PageMenu', PageMenu)
app.component('PageMeta', PageMeta)
app.component('RouterLink', RouterLink)
app.component('Binder', Binder)
app.component('PageFoot', PageFoot)
app.component('TocMain', TocMain)

app.mount('#app')

runbox()
