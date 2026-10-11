import { fileURLToPath } from 'node:url'

import { defineConfig, configDefaults } from 'vitest/config'

export default defineConfig({
  resolve: {
    alias: {
      $shared: fileURLToPath(new URL('./src/shared', import.meta.url)),
    },
  },
  test: {
    environment: 'jsdom',
    exclude: [...configDefaults.exclude, 'e2e/**'],
  },
})
