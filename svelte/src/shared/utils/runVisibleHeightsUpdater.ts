// $shared/utils/runVisibleHeightsUpdater.ts
export function runVisibleHeightsUpdater() {
  let started = false

  const calc = (el: HTMLElement | null) => {
    if (!el) return 0
    const r = el.getBoundingClientRect()
    const vh = window.innerHeight
    return Math.max(0, Math.min(r.bottom, vh) - Math.max(r.top, 0))
  }

  const update = () => {
    const nav = document.getElementById('navbar')
    const foot = document.getElementById('footer')
    const html = document.documentElement
    html.style.setProperty('--navbar-visible-height', `${calc(nav)}px`)
    html.style.setProperty('--footer-visible-height', `${calc(foot)}px`)
  }

  const start = () => {
    if (started) return
    started = true
    update()
    requestAnimationFrame(update)
    window.addEventListener('scroll', update, { passive: true })
    window.addEventListener('resize', update)

    const ro = new ResizeObserver(update)
    ro.observe(document.documentElement)
  }

  if (document.readyState === 'complete') {
    start()
    return
  }

  window.addEventListener('load', start, { once: true })
}
