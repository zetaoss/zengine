import { createApp } from 'vue'

import LayoutFoot from '@common/components/LayoutFoot.vue'
import LayoutRemocon from '@common/components/LayoutRemocon.vue'
import NavbarSearch from '@common/components/navbar/NavbarSearch.vue'
import NavbarUserMenu from '@/components/NavbarUserMenu.vue'

import PageMenu from '@/components/PageMenu.vue'
import PageMeta from '@/components/PageMeta.vue'
import RouterLink from '@/components/RouterLink.vue'
import TheBinder from '@/components/binder/TheBinder.vue'
import PageFoot from '@/components/pagefoot/PageFoot.vue'
import TocMain from '@/components/toc/TocMain.vue'

import runbox from '@/components/runbox/runbox'

import '@/assets/app.scss'

const app = createApp({})
app.component('LayoutFoot', LayoutFoot)
app.component('LayoutRemocon', LayoutRemocon)
app.component('NavbarSearch', NavbarSearch)
app.component('NavbarUserMenu', NavbarUserMenu)
app.component('PageMenu', PageMenu)
app.component('PageMeta', PageMeta)
app.component('RouterLink', RouterLink)
app.component('TheBinder', TheBinder)
app.component('PageFoot', PageFoot)
app.component('TocMain', TocMain)
app.mount('#app')

runbox()
