import adapter from '@sveltejs/adapter-static'
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte'

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  onwarn: (warning, defaultHandler) => {
    if (warning.code === 'options_missing_custom_element') return
    defaultHandler(warning)
  },
  kit: {
    alias: {
      $shared: 'src/shared',
    },
    adapter: adapter({
      pages: 'dist',
      assets: 'dist',
      fallback: 'index.html',
      strict: false,
    }),
  },
}

export default config
