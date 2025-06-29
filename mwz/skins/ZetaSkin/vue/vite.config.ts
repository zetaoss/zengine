import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  build: {
    outDir: '../resources/dist',
    emptyOutDir: true,
    // <terser>
    // Prevent conflict with MediaWiki (jQuery, VisualEditor)
    // https://github.com/evanw/esbuild/issues/2338
    minify: 'terser',
    terserOptions: {
      compress: {
        passes: 1,
      },
      mangle: {
        reserved: ['$', 've'],
      },
    },
    // </terser>
    rollupOptions: {
      input: 'src/app.ts',
      output: {
        compact: true,
        entryFileNames: '[name].js',
        assetFileNames: '[name].[ext]',
      },
    },
  },
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@common': fileURLToPath(new URL('./common', import.meta.url)),
    },
  },
})
