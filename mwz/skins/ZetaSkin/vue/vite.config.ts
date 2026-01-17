import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import { defineConfig, type Plugin } from 'vite'

function devDist(): Plugin {
  return {
    name: 'dev-dist',
    apply: 'serve',
    configureServer() {
      const distDir = path.resolve(__dirname, '../resources/dist')
      const appJsContent = `;(function () {
  var s = document.createElement('script')
  s.type = 'module'
  s.src = '/dev5174/src/app.ts?v=' + Date.now()
  document.head.appendChild(s)})()`
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
