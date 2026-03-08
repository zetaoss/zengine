const SCOPE = 'Path=/; SameSite=Lax; Secure'
const TTL = 31_536_000

export function readCookie(name: string): string {
  const key = name + '='
  for (const token of document.cookie.split(';')) {
    const cookie = token.trim()
    if (!cookie.startsWith(key)) continue
    const raw = cookie.slice(key.length)
    try {
      return decodeURIComponent(raw)
    } catch {
      // Broken percent-encoding can exist in legacy/external cookies.
      return raw
    }
  }
  return ''
}

export function writeCookie(name: string, value: string): void {
  document.cookie = `${name}=${encodeURIComponent(value)}; ${SCOPE}; Max-Age=${TTL}`
}

export function deleteCookie(name: string): void {
  document.cookie = `${name}=; ${SCOPE}; Max-Age=0`
}
