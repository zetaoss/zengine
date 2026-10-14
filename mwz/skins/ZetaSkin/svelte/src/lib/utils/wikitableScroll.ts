import { mdiChevronLeft, mdiChevronRight } from '@mdi/js'

// Wraps each content wikitable (articles and special pages) so its horizontal scroll shows which edge has
// more content (a faded edge) and, with a mouse, gets buttons to scroll by most of a screen. The table stays
// the scroll container (wikitable.css), so without this script tables still scroll. #sv-software is a card
// grid, not a scrolling table.
const EDGE_TOLERANCE = 1

function button(side: 'start' | 'end', table: HTMLElement): HTMLButtonElement {
  const el = document.createElement('button')
  el.type = 'button'
  el.className = `wikitable-scroll-btn wikitable-scroll-btn-${side}`
  el.tabIndex = -1
  el.setAttribute('aria-hidden', 'true')
  el.innerHTML = `<svg viewBox="0 0 24 24" aria-hidden="true"><path d="${side === 'start' ? mdiChevronLeft : mdiChevronRight}"/></svg>`
  el.addEventListener('click', () => {
    table.scrollBy({ left: (side === 'start' ? -1 : 1) * table.clientWidth * 0.8, behavior: 'smooth' })
  })
  return el
}

function update(wrapper: HTMLElement, table: HTMLElement): void {
  const scrolled = Math.abs(table.scrollLeft)
  const overflow = table.scrollWidth - table.clientWidth > EDGE_TOLERANCE
  wrapper.classList.toggle('more-start', overflow && scrolled > EDGE_TOLERANCE)
  wrapper.classList.toggle('more-end', overflow && scrolled + table.clientWidth < table.scrollWidth - EDGE_TOLERANCE)
}

export default function wikitableScroll(): void {
  const tables = Array.from(document.querySelectorAll<HTMLTableElement>('#mw-content-text table.wikitable:not(#sv-software)')).filter(
    (table) => !table.parentElement?.closest('table.wikitable'),
  )
  if (tables.length === 0) return

  const wrappers = new Map<Element, HTMLElement>()
  const resizeObserver = new ResizeObserver((entries) => {
    for (const entry of entries) {
      const wrapper = wrappers.get(entry.target)
      if (wrapper) update(wrapper, entry.target as HTMLElement)
    }
  })

  for (const table of tables) {
    const wrapper = document.createElement('div')
    wrapper.className = 'wikitable-scroll'
    table.before(wrapper)
    wrapper.append(button('start', table), table, button('end', table))
    wrappers.set(table, wrapper)

    let frame = 0
    table.addEventListener(
      'scroll',
      () => {
        if (frame) return
        frame = requestAnimationFrame(() => {
          frame = 0
          update(wrapper, table)
        })
      },
      { passive: true },
    )
    resizeObserver.observe(table)
  }
}
