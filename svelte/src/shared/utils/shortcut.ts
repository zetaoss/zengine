function getModifier(accesskey?: string): string {
  if (typeof navigator === 'undefined') return 'Alt'

  const ua = navigator.userAgent
  const isMac = /Macintosh|Mac OS X|iPhone|iPad|iPod/i.test(ua)
  const isFirefox = /Firefox\//i.test(ua)
  const isChrome = /Chrome\//i.test(ua) && !/Edg\//i.test(ua) && !/OPR\//i.test(ua)

  if (isMac) return 'Ctrl+Option'
  if (isChrome && accesskey && /^[def]$/.test(accesskey)) return 'Alt+Shift'
  if (isFirefox) return 'Alt+Shift'
  return 'Alt'
}

export default function getShortcut(accesskey?: string): string {
  return accesskey ? `${getModifier(accesskey)}+${accesskey.toUpperCase()}` : ''
}
