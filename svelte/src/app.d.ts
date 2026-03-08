/// <reference types="@sveltejs/kit" />
import type { ZConf } from '$shared/utils/zConf'

// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces

declare global {
  interface Window {
    __gtagInitialized__?: boolean
    __gtagLastTrackedUrl__?: string
    __gtagNavTrackerInstalled__?: boolean
    __gtagScriptPromise__?: Promise<boolean>
    __gtagTrackQueued__?: boolean
    adsbygoogle: Array<Record<string, unknown>>
    dataLayer?: unknown[]
    gtag?: (...args: unknown[]) => void
    ZCONF?: Partial<ZConf>
  }

  namespace App {
    // interface Error {}
    // interface Locals {}
    // interface PageData {}
    // interface PageState {}
    // interface Platform {}
  }
}

export {}
