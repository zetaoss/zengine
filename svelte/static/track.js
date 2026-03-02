;(function () {
  var klaroLoaded = false
  var gaLoaded = false
  var canTrack = false
  var lastTrackedUrl = ''
  var isQueued = false
  var eeaCache = null

  function getConsent(serviceName) {
    var raw = localStorage.getItem('klaro')
    if (!raw) return null

    try {
      var parsed = JSON.parse(raw)
      if (!parsed || typeof parsed !== 'object') return null

      var consents = parsed.consents
      if (!consents || typeof consents !== 'object') return null

      var consent = consents[serviceName]
      return typeof consent === 'boolean' ? consent : null
    } catch {
      return null
    }
  }

  function updateGaConsent(analyticsConsent, adsConsent) {
    var gtag = window.gtag
    if (typeof gtag !== 'function') return

    gtag('consent', 'update', {
      ad_personalization: adsConsent ? 'granted' : 'denied',
      ad_storage: adsConsent ? 'granted' : 'denied',
      ad_user_data: adsConsent ? 'granted' : 'denied',
      analytics_storage: analyticsConsent ? 'granted' : 'denied',
    })
  }

  // https://klaro.org/docs/integration/annotated-configuration
  window.klaroConfig = {
    lang: 'en',
    storageMethod: 'localStorage',
    services: [
      {
        name: 'google-analytics',
        purposes: ['analytics'],
        callback: function (consent) {
          onAnalyticsConsentChange(consent)
        },
      },
      {
        name: 'google-adsense',
        purposes: ['advertising'],
        callback: function (consent) {
          var analyticsConsent = getConsent('google-analytics') === true
          if (analyticsConsent || consent === true) {
            loadGa()
          }
          updateGaConsent(analyticsConsent, consent === true)
        },
      },
    ],
  }

  function onAnalyticsConsentChange(consent) {
    var analyticsConsent = consent === true
    var adsConsent = getConsent('google-adsense') === true

    if (analyticsConsent || adsConsent) {
      loadGa()
    }
    updateGaConsent(analyticsConsent, adsConsent)

    canTrack = analyticsConsent
    if (!canTrack) return

    trackUrl()
  }

  function loadKlaro() {
    if (klaroLoaded) return
    klaroLoaded = true

    var link = document.createElement('link')
    link.rel = 'stylesheet'
    link.href = 'https://cdn.jsdelivr.net/npm/klaro@0.7.22/dist/klaro.min.css'
    document.head.appendChild(link)

    var script = document.createElement('script')
    script.src = 'https://cdn.jsdelivr.net/npm/klaro@0.7.22/dist/klaro.min.js'
    document.head.appendChild(script)
  }

  function loadGa() {
    if (gaLoaded) return

    var measurementId = window.__CONFIG__ && window.__CONFIG__.gaMeasurementId
    if (typeof measurementId !== 'string' || measurementId.length === 0) return
    gaLoaded = true

    window.dataLayer = window.dataLayer || []
    var gtag =
      window.gtag ||
      function () {
        if (window.dataLayer) {
          window.dataLayer.push(arguments)
        }
      }
    window.gtag = gtag

    var consent = 'denied'
    gtag('consent', 'default', {
      ad_personalization: consent,
      ad_storage: consent,
      ad_user_data: consent,
      analytics_storage: consent,
    })
    gtag('js', new Date())
    gtag('config', measurementId, { send_page_view: false })

    var script = document.createElement('script')
    script.async = true
    script.src = 'https://www.googletagmanager.com/gtag/js?id=' + encodeURIComponent(measurementId)
    document.head.appendChild(script)
  }

  function sendPageView() {
    var gtag = window.gtag
    if (!gtag) return
    var location = window.location

    gtag('event', 'page_view', {
      page_path: location.pathname + location.search,
      page_location: location.href,
      page_title: document.title,
    })
  }

  function trackUrl() {
    if (canTrack !== true) return

    var href = window.location.href
    if (href === lastTrackedUrl) return
    lastTrackedUrl = href

    sendPageView()
  }

  function scheduleTrack() {
    if (isQueued) return
    isQueued = true

    queueMicrotask(function () {
      isQueued = false
      trackUrl()
    })
  }

  function installNavTracker() {
    var pushState = window.history.pushState
    var replaceState = window.history.replaceState

    window.history.pushState = function () {
      var result = pushState.apply(this, arguments)
      scheduleTrack()
      return result
    }
    window.history.replaceState = function () {
      var result = replaceState.apply(this, arguments)
      scheduleTrack()
      return result
    }
    window.addEventListener('popstate', scheduleTrack)
  }

  async function getIsEEA(skipEeaLookup) {
    if (skipEeaLookup) return false
    if (typeof eeaCache === 'boolean') return eeaCache

    var cached = localStorage.getItem('eea')
    if (cached === '1' || cached === '0') {
      eeaCache = cached === '1'
      return eeaCache
    }

    try {
      var response = await fetch('/eea/v1')
      if (!response.ok) throw new Error('bad status')

      var payload = (await response.text()).trim()
      if (payload !== '0' && payload !== '1') throw new Error('invalid payload')
      localStorage.setItem('eea', payload)
      eeaCache = payload === '1'
      return eeaCache
    } catch {
      // Fail safe: if lookup fails, assume EEA and require consent.
      eeaCache = true
      return eeaCache
    }
  }

  async function init() {
    var hasKlaro = localStorage.getItem('klaro') !== null
    var hasAnalyticsConsent = getConsent('google-analytics') === true
    var hasAdsConsent = getConsent('google-adsense') === true
    var isEEA = await getIsEEA(hasKlaro)

    if (hasKlaro || isEEA) {
      loadKlaro()
    }

    canTrack = hasAnalyticsConsent
    if (hasAnalyticsConsent || hasAdsConsent) {
      loadGa()
      updateGaConsent(hasAnalyticsConsent, hasAdsConsent)
    }
    trackUrl()
  }

  installNavTracker()
  void init()
})()
