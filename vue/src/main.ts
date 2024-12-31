import { createApp } from 'vue'

import { useDark, useToggle } from '@vueuse/core'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import './assets/app.scss'

useToggle(useDark())

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
