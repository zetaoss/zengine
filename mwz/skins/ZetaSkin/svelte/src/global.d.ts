import type { ZConf } from '$shared/utils/zConf'

export {}

declare global {
  interface Window {
    __gtagConfiguredMeasurementId__?: string
    __gtagInitialized__?: boolean
    __gtagLastTrackedUrl__?: string
    __gtagNavTrackerInstalled__?: boolean
    __gtagScriptPromise__?: Promise<boolean>
    __gtagTrackQueued__?: boolean
    adsbygoogle?: Array<Record<string, unknown>>
    dataLayer?: unknown[]
    gtag?: (...args: unknown[]) => void
    RLCONF?: {
      wgAction?: string
      wgArticleId?: number
      wgUserId?: number
    }
    ZCONF?: Partial<ZConf>
  }
}
