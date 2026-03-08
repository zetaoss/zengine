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
import '$shared/components/CAdsenseSlot.svelte'
import '$shared/components/CTrackingScripts.svelte'

import shortcutForPageButtons from '$lib/utils/shortcutForPageButtons'
import { setTrackingGate } from '$shared/stores/trackingStore'
import { runVisibleHeightsUpdater } from '$shared/utils/runVisibleHeightsUpdater'

import { mountRunbox } from './components/runbox/runbox'

setTrackingGate(() => window.RLCONF?.wgAction === 'view' && (window.RLCONF?.wgUserId ?? 0) === 0 && (window.RLCONF?.wgArticleId ?? 0) > 0)

runVisibleHeightsUpdater()
mountRunbox()
shortcutForPageButtons()
