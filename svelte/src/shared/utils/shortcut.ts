function getModifier(): string {
  if (typeof navigator === 'undefined') return 'Alt'

  const ua = navigator.userAgent
  const isMac = /Macintosh|Mac OS X|iPhone|iPad|iPod/i.test(ua)
  const isFirefox = /Firefox\//i.test(ua)

  if (isMac) return 'Ctrl+Option'
  if (isFirefox) return 'Alt+Shift'
  return 'Alt'
}

export default function getShortcut(accesskey?: string): string | undefined {
  const key = accesskey?.trim()
  if (!key) return undefined
  const display = key.length === 1 ? key.toUpperCase() : key
  return `${getModifier()}+${display}`
}
