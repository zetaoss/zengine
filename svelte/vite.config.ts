import { readFile, writeFile } from 'node:fs/promises'
import path from 'node:path'

import { sveltekit } from '@sveltejs/kit/vite'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig, type Plugin, type ResolvedConfig, transformWithEsbuild } from 'vite'

function minifyJsPlugin(): Plugin {
  let resolvedConfig: ResolvedConfig

  return {
    name: 'minify-js',
    apply: 'build',
    configResolved(config) {
      resolvedConfig = config
    },
    async writeBundle() {
      const trackPath = path.resolve(resolvedConfig.root, resolvedConfig.build.outDir, 'track.js')

      let source: string
      try {
        source = await readFile(trackPath, 'utf8')
      } catch (error) {
        if ((error as NodeJS.ErrnoException).code === 'ENOENT') return
        throw error
      }

      const minified = await transformWithEsbuild(source, trackPath, {
        minify: true,
        legalComments: 'none',
      })
      await writeFile(trackPath, minified.code)
    },
  }
}

// https://vite.dev/config/
export default defineConfig({
  plugins: [sveltekit(), tailwindcss(), minifyJsPlugin()],
  server: {
    allowedHosts: true,
    host: true,
    port: 5173,
    strictPort: true,
  },
})
