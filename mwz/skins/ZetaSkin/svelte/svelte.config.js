/** @type {import('svelte/compiler').CompileOptions} */
const compilerOptions = {
  customElement: true,
}

export default {
  compilerOptions,
  onwarn: (warning, defaultHandler) => {
    if (warning.code === 'css_unused_selector') {
      throw new Error(`[svelte:${warning.code}] ${warning.message}`)
    }
    defaultHandler(warning)
  },
}
