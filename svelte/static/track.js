;(function () {
  function bootTracking() {
    if (window.__zetaTrackingBooted__) return
    window.__zetaTrackingBooted__ = true

    function readCookie(name) {
      const key = name + '='
      const tokens = document.cookie ? document.cookie.split(';') : []
      for (let i = 0; i < tokens.length; i += 1) {
        const token = tokens[i].trim()
        if (!token.startsWith(key)) continue
        const raw = token.slice(key.length)
        try {
          return decodeURIComponent(raw)
        } catch {
          return raw
        }
      }
      return ''
    }

    function getZConf() {
      return window.ZCONF || {}
    }

    function getTrackingState() {
      const zconf = getZConf()
      const policy = zconf && zconf.policy === 'standard' ? 'standard' : 'strict'
      const rawConsent = readCookie('consent')
      const hasConsentCookie = rawConsent.length > 0

      let grants
      if (!rawConsent) {
        const granted = policy === 'standard' ? 'granted' : 'denied'
        grants = {
          analyticsStorage: granted,
          adStorage: granted,
          adUserData: granted,
          adPersonalization: granted,
        }
      } else {
        const params = new URLSearchParams(rawConsent)
        const parseValue = function (raw) {
          return raw === 'granted' ? 'granted' : 'denied'
        }
        grants = {
          analyticsStorage: parseValue(params.get('analytics_storage')),
          adStorage: parseValue(params.get('ad_storage')),
          adUserData: parseValue(params.get('ad_user_data')),
          adPersonalization: parseValue(params.get('ad_personalization')),
        }
      }

      const consentAnalytics = grants.analyticsStorage === 'granted'
      const consentAds = grants.adStorage === 'granted'
      return {
        consentAnalytics: consentAnalytics,
        consentAds: consentAds,
        grants: grants,
        hasConsentCookie: hasConsentCookie,
        policy: policy,
        canBootGtag: consentAnalytics || consentAds,
        canShowAds: consentAds,
      }
    }

    function canBootGtag(zconf) {
      const policy = zconf && zconf.policy
      const rawConsent = readCookie('consent')
      if (!rawConsent) return policy === 'standard'
      const params = new URLSearchParams(rawConsent)
      return params.get('analytics_storage') === 'granted' || params.get('ad_storage') === 'granted'
    }

    function loadGtagScript(measurementId) {
      const w = window
      if (w.__gtagScriptPromise__) return w.__gtagScriptPromise__

      const encodedId = encodeURIComponent(measurementId)
      const existing = document.querySelector(
        'script[data-zeta-gtag="' + measurementId + '"],script[src*="googletagmanager.com/gtag/js"][src*="id=' + encodedId + '"]',
      )
      if (existing) {
        if (existing.dataset.zetaLoaded === '1') {
          w.__gtagScriptPromise__ = Promise.resolve(true)
          return w.__gtagScriptPromise__
        }

        w.__gtagScriptPromise__ = new Promise(function (resolve) {
          let settled = false
          const onLoad = function () {
            existing.dataset.zetaLoaded = '1'
            finish(true)
          }
          const onError = function () {
            finish(false)
          }
          const timeout = window.setTimeout(function () {
            finish(false)
          }, 15000)
          const finish = function (ok) {
            if (settled) return
            settled = true
            clearTimeout(timeout)
            existing.removeEventListener('load', onLoad)
            existing.removeEventListener('error', onError)
            if (!ok) w.__gtagScriptPromise__ = null
            resolve(ok)
          }
          existing.addEventListener('load', onLoad, { once: true })
          existing.addEventListener('error', onError, { once: true })
        })
        return w.__gtagScriptPromise__
      }

      w.__gtagScriptPromise__ = new Promise(function (resolve) {
        const script = document.createElement('script')
        script.async = true
        script.dataset.zetaGtag = measurementId
        script.src = 'https://www.googletagmanager.com/gtag/js?id=' + encodedId
        script.addEventListener(
          'load',
          function () {
            script.dataset.zetaLoaded = '1'
            resolve(true)
          },
          { once: true },
        )
        script.addEventListener(
          'error',
          function () {
            w.__gtagScriptPromise__ = null
            resolve(false)
          },
          { once: true },
        )
        document.head.appendChild(script)
      })
      return w.__gtagScriptPromise__
    }

    function loadAdsenseScript() {
      const state = getTrackingState()
      if (!state.canShowAds) return Promise.resolve(false)

      const client = getZConf().adClient
      if (!client) return Promise.resolve(false)

      const w = window
      if (w.__adsenseScriptPromise__) return w.__adsenseScriptPromise__

      const existing = document.querySelector('script[src*="pagead2.googlesyndication.com/pagead/js/adsbygoogle.js"]')
      if (existing) {
        if (Array.isArray(window.adsbygoogle)) {
          window.dispatchEvent(new Event('adsense:ready'))
          w.__adsenseScriptPromise__ = Promise.resolve(true)
          return w.__adsenseScriptPromise__
        }

        w.__adsenseScriptPromise__ = new Promise(function (resolve) {
          existing.addEventListener(
            'load',
            function () {
              window.dispatchEvent(new Event('adsense:ready'))
              resolve(true)
            },
            { once: true },
          )
          existing.addEventListener(
            'error',
            function () {
              window.dispatchEvent(new Event('adsense:error'))
              w.__adsenseScriptPromise__ = null
              resolve(false)
            },
            { once: true },
          )
        })
        return w.__adsenseScriptPromise__
      }

      w.__adsenseScriptPromise__ = new Promise(function (resolve) {
        const script = document.createElement('script')
        script.async = true
        script.crossOrigin = 'anonymous'
        script.src = 'https://pagead2.googlesyndication.com/pagead/js/adsbygoogle.js?client=' + encodeURIComponent(client)
        script.addEventListener(
          'load',
          function () {
            window.dispatchEvent(new Event('adsense:ready'))
            resolve(true)
          },
          { once: true },
        )
        script.addEventListener(
          'error',
          function () {
            window.dispatchEvent(new Event('adsense:error'))
            w.__adsenseScriptPromise__ = null
            resolve(false)
          },
          { once: true },
        )
        document.head.appendChild(script)
      })
      return w.__adsenseScriptPromise__
    }

    function updateAnalyticsConsent() {
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

    function trackPageView() {
      const state = getTrackingState()
      if (!state.consentAnalytics) return
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

    function scheduleTrack() {
      const w = window
      if (w.__gtagTrackQueued__) return
      w.__gtagTrackQueued__ = true
      queueMicrotask(function () {
        w.__gtagTrackQueued__ = false
        trackPageView()
      })
    }

    function installNavTracker() {
      const w = window
      if (w.__gtagNavTrackerInstalled__) return
      w.__gtagNavTrackerInstalled__ = true

      const pushState = window.history.pushState
      const replaceState = window.history.replaceState

      window.history.pushState = function () {
        const result = pushState.apply(this, arguments)
        scheduleTrack()
        return result
      }
      window.history.replaceState = function () {
        const result = replaceState.apply(this, arguments)
        scheduleTrack()
        return result
      }
      window.addEventListener('popstate', scheduleTrack)
    }

    function ensureAnalytics() {
      const measurementId = (getZConf().gaMeasurementId || '').trim()
      const state = getTrackingState()
      if (!measurementId || !state.canBootGtag) return Promise.resolve()

      installNavTracker()

      const w = window
      w.dataLayer = w.dataLayer || []
      w.gtag =
        w.gtag ||
        function () {
          w.dataLayer.push(arguments)
        }

      if (w.__gtagInitialized__ && w.__gtagConfiguredMeasurementId__ === measurementId) {
        updateAnalyticsConsent()
        trackPageView()
        return Promise.resolve()
      }

      return loadGtagScript(measurementId).then(function (loaded) {
        if (!loaded) return
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
      })
    }

    function refreshTracking() {
      void loadAdsenseScript()
      void ensureAnalytics()
    }

    refreshTracking()
    document.addEventListener('visibilitychange', scheduleTrack)
    window.addEventListener('consent:updated', refreshTracking)
    if (canBootGtag(getZConf())) scheduleTrack()
  }

  bootTracking()
})()
