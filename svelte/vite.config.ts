import { readFile, writeFile } from 'node:fs/promises'
import { resolve } from 'node:path'

import { sveltekit } from '@sveltejs/kit/vite'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig, type Plugin, transformWithEsbuild } from 'vite'

const minifyStaticJsPlugin = (): Plugin => ({
  name: 'minify-static-js',
  apply: 'build',
  enforce: 'post',
  async closeBundle() {
    const root = process.cwd()
    const clientOutputDir = resolve(root, '.svelte-kit/output/client')
    const paths = ['track.js']

    await Promise.all(
      paths.map(async (relPath) => {
        const outputPath = resolve(clientOutputDir, relPath)
        let source: string
        try {
          source = await readFile(outputPath, 'utf8')
        } catch {
          console.warn(`[minify-static-js] Could not read file to minify: ${outputPath}`)
          return
        }

        const result = await transformWithEsbuild(source, relPath, {
          minify: true,
          legalComments: 'none',
          target: 'es2019',
        })
        await writeFile(outputPath, result.code, 'utf8')
      }),
    )
  },
})

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
