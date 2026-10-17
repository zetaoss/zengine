import { fileURLToPath } from 'node:url'

import { configDefaults, defineConfig } from 'vitest/config'

export default defineConfig({
  resolve: {
    alias: {
      $lib: fileURLToPath(new URL('./src/lib', import.meta.url)),
      $shared: fileURLToPath(new URL('./src/shared', import.meta.url)),
    },
  },
  test: {
    environment: 'jsdom',
    // svelte/ runs the tests under src/shared.
    exclude: [...configDefaults.exclude, 'src/shared/**'],
  },
})
