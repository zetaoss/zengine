import './assets/main.scss'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import { useDark, useToggle } from '@vueuse/core'

import App from './App.vue'
import router from './router'

useToggle(useDark())

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')
