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
let compatibilityInitialized = false
let retryScheduled = false

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

function queryMatching(root: ParentNode, selector: string): HTMLElement[] {
  const matched = Array.from(root.querySelectorAll<HTMLElement>(selector))
  if (root instanceof HTMLElement && root.matches(selector)) {
    matched.unshift(root)
  }
  return matched
}

function setupNavigationFramesInRoot(root: ParentNode): void {
  queryMatching(root, 'div.NavFrame:not(.mw-collapsible)').forEach((frame) => {
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

function setupTabsInRoot(root: ParentNode): void {
  queryMatching(root, '.zw-tabs').forEach((tabsRoot, tabIndex) => {
    if (tabsRoot.dataset.tabsReady) return
    const tabs = Array.from(tabsRoot.querySelectorAll<HTMLElement>(':scope > ul.nav > li > .nav-link'))
    const panes = Array.from(tabsRoot.querySelectorAll<HTMLElement>(':scope > .tab-content > .tab-pane'))
    if (!tabs.length || panes.length < tabs.length) return

    tabsRoot.setAttribute('data-tabs-ready', 'true')
    panes.forEach((pane, paneIndex) => {
      pane.hidden = paneIndex !== 0
      pane.classList.toggle('show', paneIndex === 0)
      pane.classList.toggle('active', paneIndex === 0)
    })

    tabs.forEach((tab, index) => {
      const id = `mytab-${tabIndex}-${index}`
      tab.dataset.target = `#${id}`
      tab.setAttribute('role', 'tab')
      tab.setAttribute('aria-selected', String(index === 0))
      tab.tabIndex = index === 0 ? 0 : -1
      panes[index].id = id
      panes[index].setAttribute('role', 'tabpanel')

      const activate = () => {
        tabs.forEach((item, itemIndex) => {
          const active = itemIndex === index
          item.classList.toggle('active', active)
          item.setAttribute('aria-selected', String(active))
          item.tabIndex = active ? 0 : -1
        })
        panes.forEach((pane, paneIndex) => {
          const active = paneIndex === index
          pane.hidden = !active
          pane.classList.toggle('show', active)
          pane.classList.toggle('active', active)
        })
      }

      tab.addEventListener('click', (event: MouseEvent) => {
        event.preventDefault()
        activate()
      })
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
  if (document.body) {
    setupNavigationFramesInRoot(document.body)
    setupTabsInRoot(document.body)
  }

  if (compatibilityInitialized) return
  const $ = runtime.jQuery
  const mw = runtime.mw
  if (typeof $ !== 'function' || typeof mw?.hook !== 'function' || typeof mw?.loader?.using !== 'function') {
    if (!retryScheduled) {
      retryScheduled = true
      window.setTimeout(() => {
        retryScheduled = false
        setupWikiCompatibility()
      }, 50)
    }
    return
  }
  compatibilityInitialized = true

  const setup = (content: JQuerySelection) => {
    setupLegacyCollapsible(content)
    content.get().forEach((element) => {
      setupNavigationFramesInRoot(element)
      setupTabsInRoot(element)
    })
  }

  mw.hook('wikipage.content').add(setup)
  mw.hook('wikipage.collapsibleContent').add((content) => setupAutocollapse(content, $))
  void mw.loader.using('jquery.makeCollapsible').then(() => setup($(document.body)))
}
