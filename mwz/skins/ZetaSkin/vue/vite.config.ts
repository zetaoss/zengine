// mwz/skins/ZetaSkin/vue/vite.config.ts
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import { defineConfig, type Plugin } from 'vite'

function devDist(): Plugin {
  let base = '/'

  return {
    name: 'dev-dist',
    apply: 'serve',
    configResolved(config) {
      base = config.base
    },
    configureServer() {
      const distDir = path.resolve(__dirname, '../resources/dist')
      const appSrc = `${base}src/app.ts?v=${Date.now()}`
      const appJsContent = `;(() => {
  const s = document.createElement('script')
  s.type = 'module'
  s.src = '${appSrc}'
  document.head.appendChild(s)
})()`
      fs.mkdirSync(distDir, { recursive: true });
      fs.writeFileSync(path.join(distDir, 'app.css'), '/* dev */', 'utf-8')
      fs.writeFileSync(path.join(distDir, 'app.js'), appJsContent, 'utf-8')
    },
  }
}

// https://vite.dev/config/
export default defineConfig({
  // dev
  base: '/dev5174/',
  server: {
    allowedHosts: true,
    host: true,
    port: 5174,
    strictPort: true,
  },
  // build
  build: {
    outDir: '../resources/dist',
    emptyOutDir: true,
    lib: {
      entry: 'src/app.ts',
      name: 'app',
      formats: ['iife'],
      fileName: () => 'app.js',
      cssFileName: 'app',
    },
  },
  plugins: [
    vue(),
    devDist(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@common': fileURLToPath(new URL('./common', import.meta.url)),
      vue: 'vue/dist/vue.esm-bundler.js',
    },
  },
})
