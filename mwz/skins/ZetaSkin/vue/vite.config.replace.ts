import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  build: {
    outDir: '../resources/dist',
    emptyOutDir: true,
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
    // <replace>
    // Prevent conflict with MediaWiki (jQuery, VisualEditor)
    // https://github.com/evanw/esbuild/issues/2338
    // Faster but unstable replace plugin for development ⚠️ Use only in dev.
    {
      name: 'replace',
      apply: 'build',
      generateBundle(_, bundle) {
        for (const [file, chunk] of Object.entries(bundle)) {
          if ('code' in chunk) { // js file only
            console.log(`Processing JS file: ${file}`);
            chunk.code = chunk.code.replace(/\bve\b/g, '___ve'); // ve → ___ve
          }
        }
      },
    },
    // </replace>
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@common': fileURLToPath(new URL('./common', import.meta.url)),
      vue: 'vue/dist/vue.esm-bundler.js',
    },
  },
})
