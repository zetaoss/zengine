import { createHash } from 'node:crypto'
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
    async closeBundle() {
      const trackPath = path.resolve(resolvedConfig.root, 'dist', 'track.js')
      const indexPath = path.resolve(resolvedConfig.root, 'dist', 'index.html')

      let trackSource: string
      try {
        trackSource = await readFile(trackPath, 'utf8')
      } catch (error) {
        if ((error as NodeJS.ErrnoException).code === 'ENOENT') return
        throw error
      }

      const trackVersion = createHash('sha256').update(trackSource).digest('hex').slice(0, 12)
      let indexHtml: string
      try {
        indexHtml = await readFile(indexPath, 'utf8')
      } catch (error) {
        if ((error as NodeJS.ErrnoException).code === 'ENOENT') return
        throw error
      }

      await writeFile(
        indexPath,
        indexHtml.replace('<script src="/track.js" defer></script>', `<script src="/track.js?v=${trackVersion}" defer></script>`),
      )
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
