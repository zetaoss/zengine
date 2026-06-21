const customElementFiles = [
  'src/components/PageFooter.svelte',
  'src/components/PageMenu.svelte',
  'src/components/PageMeta.svelte',
  'src/components/UserMenu.svelte',
  'src/components/binder/BinderApex.svelte',
  'src/components/editbox/EditBox.svelte',
  'src/components/toc/TocApex.svelte',
  'src/shared/components/CAdsenseSlot.svelte',
  'src/shared/components/CSiteFooter.svelte',
  'src/shared/components/CSiteRemocon.svelte',
  'src/shared/components/navbar/CSearch.svelte',
  'src/shared/ui/CBadge.svelte',
  'src/shared/ui/confirm/ConfirmHost.svelte',
  'src/shared/ui/toast/ToastHost.svelte',
]

/** @param {{ filename: string }} options */
export function isCustomElement({ filename }) {
  const normalizedFilename = filename.replaceAll('\\', '/')
  return customElementFiles.some((file) => normalizedFilename.endsWith(file))
}

/** @type {import('svelte/compiler').CompileOptions} */
const compilerOptions = {
  customElement: isCustomElement,
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
