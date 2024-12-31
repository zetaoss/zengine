import path from 'path'

import vue from '@vitejs/plugin-vue'
import laravel from 'laravel-vite-plugin'
import { defineConfig } from 'vite'

export default defineConfig({
  clearScreen: false,
  build: {
    emptyOutDir: false,
    outDir: 'resources/dist',
    rollupOptions: {
      output: {
        compact: true,
        entryFileNames: '[name].js',
        assetFileNames: '[name].[ext]',
      },
    },
    // to avoid global name conflict https://github.com/vitejs/vite/issues/5426#issuecomment-1085684107
    minify: 'terser',
    terserOptions: {
      mangle: {
        reserved: ['$', 'mw', 've'],
      },
    },
  },
  plugins: [
    laravel({
      input: [
        'src/app.ts',
      ],
      refresh: true,
    }),
    vue(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@common': path.resolve(__dirname, './common'),
      vue: 'vue/dist/vue.esm-bundler.js',
    },
  },
  server: {
    watch: {
      usePolling: true,
      interval: 100,
    },
  },
})
