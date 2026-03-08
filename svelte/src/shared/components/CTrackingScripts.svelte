<svelte:options customElement={{ tag: 'c-tracking-scripts', shadow: 'none' }} />

<script lang="ts">
  import { onMount } from 'svelte'

  import { getTrackingState } from '$shared/stores/trackingStore'
  import { getZConf } from '$shared/utils/zConf'

  let adsenseScriptPromise: Promise<boolean> | null = null

  const hasGtagRuntimeSignals = (): boolean => {
    const dl = window.dataLayer
    if (!Array.isArray(dl)) return false
    return dl.some((entry) => {
      if (!entry || typeof entry !== 'object' || Array.isArray(entry)) return false
      const event = (entry as { event?: unknown }).event
      return event === 'gtm.js' || event === 'gtm.dom' || event === 'gtm.load'
    })
  }

  const loadAdsenseScript = async (): Promise<boolean> => {
    if (!getTrackingState().canShowAds) return false

    const client = getZConf().adClient
    if (!client) return false

    if (adsenseScriptPromise) return adsenseScriptPromise

    const existing = document.querySelector<HTMLScriptElement>('script[src*="pagead2.googlesyndication.com/pagead/js/adsbygoogle.js"]')
    if (existing) {
      if (Array.isArray(window.adsbygoogle)) {
        window.dispatchEvent(new Event('adsense:ready'))
        adsenseScriptPromise = Promise.resolve(true)
        return adsenseScriptPromise
      }

      adsenseScriptPromise = new Promise<boolean>((resolve) => {
        existing.addEventListener(
          'load',
          () => {
            window.dispatchEvent(new Event('adsense:ready'))
            resolve(true)
          },
          { once: true },
        )
        existing.addEventListener(
          'error',
          () => {
            window.dispatchEvent(new Event('adsense:error'))
            resolve(false)
          },
          { once: true },
        )
      })
      return adsenseScriptPromise
    }

    adsenseScriptPromise = new Promise<boolean>((resolve) => {
      const script = document.createElement('script')
      script.async = true
      script.crossOrigin = 'anonymous'
      script.src = `https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=${encodeURIComponent(client)}`
      script.addEventListener(
        'load',
        () => {
          window.dispatchEvent(new Event('adsense:ready'))
          resolve(true)
        },
        { once: true },
      )
      script.addEventListener(
        'error',
        () => {
          window.dispatchEvent(new Event('adsense:error'))
          resolve(false)
        },
        { once: true },
      )
      document.head.appendChild(script)
    })

    return adsenseScriptPromise
  }

  const loadGtagScript = (measurementId: string): Promise<boolean> => {
    const w = window
    if (w.__gtagScriptPromise__) return w.__gtagScriptPromise__

    const existing = document.querySelector<HTMLScriptElement>(`script[src*="googletagmanager.com/gtag/js"][src*="id=${measurementId}"]`)
    if (existing) {
      w.__gtagScriptPromise__ = new Promise<boolean>((resolve) => {
        if (existing.dataset.zetaLoaded === '1' || hasGtagRuntimeSignals()) {
          resolve(true)
          return
        }

        const onLoad = () => {
          existing.dataset.zetaLoaded = '1'
          resolve(true)
        }
        const onError = () => resolve(false)
        existing.addEventListener('load', onLoad, { once: true })
        existing.addEventListener('error', onError, { once: true })
        window.setTimeout(() => resolve(hasGtagRuntimeSignals()), 1500)
      })
      return w.__gtagScriptPromise__
    }

    w.__gtagScriptPromise__ = new Promise<boolean>((resolve) => {
      const script = document.createElement('script')
      script.async = true
      script.dataset.zetaGtag = measurementId
      script.src = `https://www.googletagmanager.com/gtag/js?id=${encodeURIComponent(measurementId)}`
      script.onload = () => {
        script.dataset.zetaLoaded = '1'
        resolve(true)
      }
      script.onerror = () => resolve(false)
      document.head.appendChild(script)
    })
    return w.__gtagScriptPromise__
  }

  const updateAnalyticsConsent = (): void => {
    const gtag = window.gtag
    if (typeof gtag !== 'function') return

    const state = getTrackingState()
    gtag('consent', 'update', {
      ad_storage: state.grants.adStorage,
      ad_user_data: state.grants.adUserData,
      ad_personalization: state.grants.adPersonalization,
      analytics_storage: state.grants.analyticsStorage,
    })
  }

  const trackPageView = (): void => {
    if (!getTrackingState().consentAnalytics) return

    const w = window
    const gtag = w.gtag
    if (typeof gtag !== 'function') return
    const measurementId = w.__gtagConfiguredMeasurementId__
    if (!measurementId) return

    const href = window.location.href
    if (w.__gtagLastTrackedUrl__ === href) return
    w.__gtagLastTrackedUrl__ = href

    gtag('event', 'page_view', {
      send_to: measurementId,
      page_path: window.location.pathname + window.location.search,
      page_location: href,
      page_title: document.title,
    })
  }

  const scheduleTrack = (): void => {
    const w = window
    if (w.__gtagTrackQueued__) return
    w.__gtagTrackQueued__ = true
    queueMicrotask(() => {
      w.__gtagTrackQueued__ = false
      trackPageView()
    })
  }

  const installNavTracker = (): void => {
    const w = window
    if (w.__gtagNavTrackerInstalled__) return
    w.__gtagNavTrackerInstalled__ = true

    const pushState = window.history.pushState
    const replaceState = window.history.replaceState

    window.history.pushState = function (...args) {
      const result = pushState.apply(this, args)
      scheduleTrack()
      return result
    }

    window.history.replaceState = function (...args) {
      const result = replaceState.apply(this, args)
      scheduleTrack()
      return result
    }

    window.addEventListener('popstate', scheduleTrack)
  }

  const ensureAnalytics = async (): Promise<void> => {
    const measurementId = getZConf().gaMeasurementId
    const state = getTrackingState()
    if (!measurementId || !state.canBootGtag) return

    installNavTracker()

    const w = window
    if (w.__gtagInitialized__ && w.__gtagConfiguredMeasurementId__ === measurementId) {
      updateAnalyticsConsent()
      trackPageView()
      return
    }

    const loaded = await loadGtagScript(measurementId)
    if (!loaded) return

    w.dataLayer = w.dataLayer || []
    w.gtag = w.gtag || ((...args: unknown[]) => w.dataLayer?.push(args))

    if (w.__gtagInitialized__ && w.__gtagConfiguredMeasurementId__ === measurementId) {
      updateAnalyticsConsent()
      trackPageView()
      return
    }

    w.gtag('consent', 'default', {
      ad_storage: 'denied',
      ad_user_data: 'denied',
      ad_personalization: 'denied',
      analytics_storage: 'denied',
    })
    w.gtag('js', new Date())
    w.gtag('config', measurementId, { send_page_view: false })
    w.__gtagInitialized__ = true
    w.__gtagConfiguredMeasurementId__ = measurementId

    updateAnalyticsConsent()
    trackPageView()
  }

  const refreshTracking = (): void => {
    void loadAdsenseScript()
    void ensureAnalytics()
  }

  onMount(() => {
    refreshTracking()
    document.addEventListener('visibilitychange', scheduleTrack)
    window.addEventListener('consent:updated', refreshTracking)
    return () => {
      document.removeEventListener('visibilitychange', scheduleTrack)
      window.removeEventListener('consent:updated', refreshTracking)
    }
  })
</script>
