import { createHash } from 'node:crypto'
import fs from 'node:fs'
import path from 'node:path'

import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig, type Plugin } from 'vite'

import { isCustomElement } from './svelte.config.js'

const outDir = '../dist'
const distDir = path.resolve(__dirname, outDir)
const trackPath = path.resolve(__dirname, '../../../../svelte/dist/track.js')

const contentHash = (file: string) => createHash('sha256').update(fs.readFileSync(file)).digest('hex').slice(0, 8)

// dist/assets.php gives SkinZetaSkin the asset URLs: content-hashed file names after a build, the dev stubs
// with their content hash as the query string under the dev server.
function writeAssets(js: string, css: string) {
  const track = contentHash(fs.existsSync(trackPath) ? trackPath : path.join(distDir, js.split('?')[0]))
  fs.writeFileSync(
    path.join(distDir, 'assets.php'),
    `<?php\n\nreturn ['js' => '${js}', 'css' => '${css}', 'track' => '${track}'];\n`,
    'utf-8',
  )
}

function devDist(): Plugin {
  let base = '/'

  return {
    name: 'dev-dist',
    apply: 'serve',
    configResolved(config) {
      base = config.base
    },
    configureServer() {
      const appSrc = `${base}src/main.ts?v=${Date.now()}`
      const appJsContent = `;(() => {
  const s = document.createElement('script')
  s.type = 'module'
  s.src = '${appSrc}'
  document.head.appendChild(s)
})()`
      fs.mkdirSync(distDir, { recursive: true })
      fs.writeFileSync(path.join(distDir, 'app.css'), '/* dev */', 'utf-8')
      fs.writeFileSync(path.join(distDir, 'app.js'), appJsContent, 'utf-8')
      writeAssets(`app.js?${contentHash(path.join(distDir, 'app.js'))}`, `app.css?${contentHash(path.join(distDir, 'app.css'))}`)
    },
  }
}

// Renames app.js/app.css to content-hashed names so each file can be cached as immutable and a new release
// never reuses an old URL.
function hashedAssets(): Plugin {
  return {
    name: 'hashed-assets',
    apply: 'build',
    closeBundle() {
      for (const f of fs.readdirSync(distDir)) {
        if (/^app\.[0-9a-f]{8}\.(js|css)$/.test(f)) fs.rmSync(path.join(distDir, f))
      }
      const names: Record<string, string> = {}
      for (const ext of ['js', 'css']) {
        const src = path.join(distDir, `app.${ext}`)
        names[ext] = `app.${contentHash(src)}.${ext}`
        fs.renameSync(src, path.join(distDir, names[ext]))
      }
      writeAssets(names.js, names.css)
    },
  }
}

export default defineConfig(({ command }) => ({
  base: '/dev5174/',
  server: {
    allowedHosts: true,
    host: true,
    port: 5174,
    strictPort: true,
    watch: {
      usePolling: true,
    },
    fs: {
      allow: ['.', path.resolve(__dirname, '../../../../svelte')],
    },
  },
  build: {
    outDir: outDir,
    emptyOutDir: false,
    lib: {
      entry: 'src/main.ts',
      name: 'app',
      formats: ['iife'],
      fileName: () => 'app.js',
      cssFileName: 'app',
    },
  },
  esbuild: {
    tsconfigRaw: {
      compilerOptions: {
        target: 'ES2022',
        module: 'ESNext',
        moduleResolution: 'Bundler',
      },
    },
  },
  define: {
    'process.env.NODE_ENV': JSON.stringify(command === 'build' ? 'production' : 'development'),
  },
  plugins: [
    svelte({
      compilerOptions: { customElement: isCustomElement },
      include: ['src/**/*.svelte'],
    }),
    tailwindcss(),
    devDist(),
    hashedAssets(),
  ],
  resolve: {
    alias: {
      $lib: '/src/lib',
      $shared: '/src/shared',
    },
    preserveSymlinks: true,
  },
}))
