import { writable } from 'svelte/store'

export const MEDIUM_SCREEN_QUERY = '(min-width: 48rem)'

export function isMdOrLarger(): boolean {
  if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return false
  return window.matchMedia(MEDIUM_SCREEN_QUERY).matches
}

export const isMdOrLargerStore = writable(isMdOrLarger())

if (typeof window !== 'undefined' && typeof window.matchMedia === 'function') {
  const media = window.matchMedia(MEDIUM_SCREEN_QUERY)
  media.addEventListener('change', (e) => isMdOrLargerStore.set(e.matches))
}
