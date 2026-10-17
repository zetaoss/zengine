interface CollapsibleController {
  collapse(): void
}

interface JQuerySelection {
  addClass(name: string): JQuerySelection
  data(name: string): CollapsibleController | undefined
  each(callback: (index: number, element: HTMLElement) => void): JQuerySelection
  find(selector: string): JQuerySelection
  get(): HTMLElement[]
  makeCollapsible(): JQuerySelection
  parent(): JQuerySelection
  hasClass(name: string): boolean
}

interface MediaWikiRuntime {
  hook(name: string): {
    add(callback: (content: JQuerySelection) => void): void
  }
  loader: {
    using(module: string): Promise<unknown>
  }
}

type JQueryFactory = (target: HTMLElement | HTMLElement[]) => JQuerySelection

const runtime = window as Window & { jQuery?: JQueryFactory; mw?: MediaWikiRuntime }

function setupLegacyCollapsible(content: JQuerySelection): void {
  const tables = content.find('table.collapsible:not(.mw-collapsible)').addClass('mw-collapsible')
  tables.each((_index, table) => {
    if (table.classList.contains('collapsed')) table.classList.add('mw-collapsed')
  })
  if (tables.get().length) {
    void runtime.mw?.loader.using('jquery.makeCollapsible').then(() => tables.makeCollapsible())
  }
}

function setupAutocollapse(content: JQuerySelection, $: JQueryFactory): void {
  const all = content.get()
  all.forEach((element) => {
    const controller = $(element).data('mw-collapsible')
    if (!controller) return

    if (all.length >= 2 && element.classList.contains('autocollapse')) {
      controller.collapse()
    } else if (element.classList.contains('innercollapse') && element.closest('.outercollapse')) {
      controller.collapse()
    }

    if (element.querySelector('.mw-collapsible-toggle') && element.style.color) {
      Array.from(element.querySelectorAll<HTMLElement>('.mw-collapsible-toggle a')).forEach((link) => {
        link.style.color = 'inherit'
      })
    }
  })
}

function setupNavigationFrames(content: JQuerySelection): void {
  content.find('div.NavFrame:not(.mw-collapsible)').each((_index, frame) => {
    if (frame.dataset.legacyNavFrameReady) return
    const head = frame.querySelector<HTMLElement>('.NavHead')
    if (!head) return
    frame.dataset.legacyNavFrameReady = 'true'

    const contentNodes = Array.from(frame.querySelectorAll<HTMLElement>(':scope > .NavPic, :scope > .NavContent'))
    const collapsed = frame.classList.contains('collapsed') || contentNodes.some((node) => node.style.display === 'none')
    const toggle = document.createElement('button')
    toggle.type = 'button'
    toggle.className = 'NavToggle'
    toggle.textContent = collapsed ? '[보이기]' : '[숨기기]'
    toggle.setAttribute('aria-expanded', String(!collapsed))
    if (head.style.color) toggle.style.color = head.style.color

    const setCollapsed = (next: boolean) => {
      contentNodes.forEach((node) => {
        if (next) {
          node.hidden = true
        } else {
          node.hidden = false
          node.style.display = ''
        }
      })
      toggle.textContent = next ? '[보이기]' : '[숨기기]'
      toggle.setAttribute('aria-expanded', String(!next))
    }

    toggle.addEventListener('click', () => setCollapsed(toggle.getAttribute('aria-expanded') === 'true'))
    head.append(toggle)
    setCollapsed(collapsed)
  })
}

function setupTabs(content: JQuerySelection): void {
  content.find('.zw-tabs').each((tabIndex, root) => {
    if (root.dataset.tabsReady) return
    const tabs = Array.from(root.querySelectorAll<HTMLElement>(':scope > ul.nav > li > .nav-link'))
    const panes = Array.from(root.querySelectorAll<HTMLElement>(':scope > .tab-content > .tab-pane'))
    if (!tabs.length || tabs.length !== panes.length) return

    root.setAttribute('data-tabs-ready', 'true')
    tabs.forEach((tab, index) => {
      const id = `mytab-${tabIndex}-${index}`
      tab.dataset.target = `#${id}`
      tab.setAttribute('role', 'tab')
      tab.setAttribute('aria-selected', String(index === 0))
      tab.tabIndex = index === 0 ? 0 : -1
      panes[index].id = id
      panes[index].setAttribute('role', 'tabpanel')
      panes[index].hidden = index !== 0

      const activate = () => {
        tabs.forEach((item, itemIndex) => {
          const active = itemIndex === index
          item.classList.toggle('active', active)
          item.setAttribute('aria-selected', String(active))
          item.tabIndex = active ? 0 : -1
          panes[itemIndex].hidden = !active
          panes[itemIndex].classList.toggle('show', active)
          panes[itemIndex].classList.toggle('active', active)
        })
      }

      tab.addEventListener('click', activate)
      tab.addEventListener('keydown', (event: KeyboardEvent) => {
        if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return
        event.preventDefault()
        const next = (index + (event.key === 'ArrowRight' ? 1 : -1) + tabs.length) % tabs.length
        tabs[next].focus()
        tabs[next].click()
      })
    })
  })
}

export default function setupWikiCompatibility(): void {
  const $ = runtime.jQuery
  const mw = runtime.mw
  if (!$ || !mw) return

  const setup = (content: JQuerySelection) => {
    setupLegacyCollapsible(content)
    setupNavigationFrames(content)
    setupTabs(content)
  }

  mw.hook('wikipage.content').add(setup)
  mw.hook('wikipage.collapsibleContent').add((content) => setupAutocollapse(content, $))
  void mw.loader.using('jquery.makeCollapsible').then(() => setup($(document.body)))
}
