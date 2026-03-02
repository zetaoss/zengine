import { createHash } from 'node:crypto'
import { readFile, writeFile } from 'node:fs/promises'
import { resolve } from 'node:path'

import { sveltekit } from '@sveltejs/kit/vite'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig, type Plugin, transformWithEsbuild } from 'vite'

const minifyStaticJsPlugin = (): Plugin => {
  return {
    name: 'minify-static-js',
    apply: 'build',
    enforce: 'post',
    async closeBundle() {
      const root = process.cwd()
      const trackPath = resolve(root, '.svelte-kit/output/client/track.js')
      const indexPath = resolve(root, 'dist/index.html')
      try {
        const source = await readFile(trackPath, 'utf8')
        const result = await transformWithEsbuild(source, 'track.js', {
          minify: true,
          legalComments: 'none',
          target: 'es2019',
        })
        await writeFile(trackPath, result.code, 'utf8')

        const trackAssetHash = createHash('sha256').update(result.code).digest('hex').slice(0, 12)
        const html = await readFile(indexPath, 'utf8')
        const patched = html.replace(/\/track\.js(?:\?v=[^"']*)?/g, `/track.js?v=${trackAssetHash}`)
        if (patched !== html) await writeFile(indexPath, patched, 'utf8')
      } catch {
        console.warn(`[minify-static-js] Could not process track.js cache-busting: ${trackPath}`)
      }
    },
  }
}

// https://vite.dev/config/
export default defineConfig({
  plugins: [sveltekit(), tailwindcss(), minifyStaticJsPlugin()],
  server: {
    allowedHosts: true,
    host: true,
    port: 5173,
    strictPort: true,
  },
})
