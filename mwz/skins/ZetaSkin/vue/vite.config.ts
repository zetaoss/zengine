import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import laravel from 'laravel-vite-plugin'

// https://vite.dev/config/
export default defineConfig({
  build: {
    emptyOutDir: true,
    outDir: '../resources/dist',
    minify: 'terser',
    terserOptions: {
      mangle: {
        reserved: ['$', 've'] // Prevent conflict with MediaWiki (jQuery, VisualEditor)
      },
    },
    rollupOptions: {
      output: {
        compact: true,
        entryFileNames: '[name].js',
        assetFileNames: '[name].[ext]',
      },
    },
  },
  plugins: [
    vue(),
    laravel({
      input: [
        'src/app.ts',
      ],
      refresh: true,
    }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@common': fileURLToPath(new URL('./common', import.meta.url)),
      vue: 'vue/dist/vue.esm-bundler.js',
    },
  },
})
