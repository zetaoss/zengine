import { URL, fileURLToPath } from 'node:url'

import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'
import FullReload from 'vite-plugin-full-reload'

// https://vitejs.dev/config/
export default defineConfig({
  clearScreen: false,
  plugins: [
    vue(),
    FullReload('../mediawiki/ZetaSkin/resources/dist/*'),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@common': fileURLToPath(new URL('./common', import.meta.url)),
    },
  },
  server: {
    hmr: {
      protocol: 'wss',
    },
    host: true,
  },
  test: {
    globals: true,
    environment: 'happy-dom',
  },
})
