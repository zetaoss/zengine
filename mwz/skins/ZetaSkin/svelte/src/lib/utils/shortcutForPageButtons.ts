import getShortcut from '$shared/utils/shortcut'

const ACCESSKEY_TARGETS = {
  'ca-edit': 'e',
  't-whatlinkshere': 'j',
  'ca-talk': 't',
} as const

export default function shortcutForPageButtons(): void {
  for (const [id, accesskey] of Object.entries(ACCESSKEY_TARGETS)) {
    const el = document.getElementById(id)
    if (!el) continue
    el.setAttribute('accesskey', accesskey)

    const shortcut = getShortcut(accesskey)
    const title = el.getAttribute('title')
    el.setAttribute('title', `${title} (${shortcut})`)
  }
}
