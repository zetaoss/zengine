// @common/composables/useVariableUpdater.ts
import { useEventListener,useWindowSize } from '@vueuse/core'
import { onMounted } from 'vue'

export function useVariableUpdater() {
  console.log('useVariableUpdater')
  const { height: winH } = useWindowSize()

  const calcVisible = (el: HTMLElement | null) => {
    if (!el) return 0
    const rect = el.getBoundingClientRect()
    return Math.max(0, Math.min(rect.bottom, winH.value) - Math.max(rect.top, 0))
  }

  const updateVars = () => {
    const navbar = document.getElementById('navbar')
    const footer = document.getElementById('footer')
    const html = document.documentElement
    html.style.setProperty('--navbar-visible-height', `${calcVisible(navbar)}px`)
    html.style.setProperty('--footer-visible-height', `${calcVisible(footer)}px`)
  }

  onMounted(() => {
    updateVars()
    useEventListener(window, 'scroll', updateVars, { passive: true })
    useEventListener(window, 'resize', updateVars)
  })
}
