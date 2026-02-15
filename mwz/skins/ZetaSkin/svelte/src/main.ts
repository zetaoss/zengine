import './assets/app.css'
import '$shared/components/navbar/CSearch.svelte'
import '$shared/components/CSiteFooter.svelte'
import '$shared/components/CSiteRemocon.svelte'
import './components/PageMeta.svelte'
import './components/PageMenu.svelte'
import './components/PageFooter.svelte'
import './components/UserMenu.svelte'
import './components/binder/BinderApex.svelte'
import './components/toc/TocApex.svelte'
import '$shared/components/CAds.svelte'

import { runVisibleHeightsUpdater } from '$shared/utils/runVisibleHeightsUpdater'

import { mountRunbox } from './components/runbox/runbox'

runVisibleHeightsUpdater()
mountRunbox()

// Preprocess light DOM content for custom elements that need slot-like HTML.
document.querySelectorAll<HTMLElement>('page-menu').forEach((node) => {
  const inner = node.innerHTML
  node.innerHTML = ''
  if (inner.trim().length) {
    node.setAttribute('slot-html', inner)
  }
})
