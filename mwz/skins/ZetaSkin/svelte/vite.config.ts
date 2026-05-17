import fs from 'node:fs'
import path from 'node:path'

import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig, type Plugin } from 'vite'

const outDir = '../dist'

function devDist(): Plugin {
  let base = '/'

  return {
    name: 'dev-dist',
    apply: 'serve',
    configResolved(config) {
      base = config.base
    },
    configureServer() {
      const distDir = path.resolve(__dirname, outDir)
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
      compilerOptions: { customElement: true },
      include: ['src/**/*.svelte'],
    }),
    tailwindcss(),
    devDist(),
  ],
  resolve: {
    alias: {
      $lib: '/src/lib',
      $shared: '/src/shared',
    },
    preserveSymlinks: true,
  },
}))
