/* eslint-disable no-undef */
/* eslint-disable @typescript-eslint/no-unused-expressions */

;(function () {
  var klaroLoaded = false
  var gaLoaded = false
  var canTrack = false
  var lastTrackedUrl = ''
  var isQueued = false

  // https://klaro.org/docs/integration/annotated-configuration
  window.klaroConfig = {
    lang: 'en',
    storageMethod: 'localStorage',
    services: [
      {
        name: 'google-analytics',
        purposes: ['analytics'],
        callback: function (consent) {
          typeof gtag == 'function' && gtag('consent', 'update', { analytics_storage: consent ? 'granted' : 'denied' })
        },
      },
      {
        name: 'google-adsense',
        purposes: ['advertising'],
        callback: function (consent) {
          typeof gtag == 'function' &&
            gtag('consent', 'update', {
              ad_storage: consent ? 'granted' : 'denied',
              ad_user_data: consent ? 'granted' : 'denied',
              ad_personalization: consent ? 'granted' : 'denied',
            })
        },
      },
    ],
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
        window.dataLayer && window.dataLayer.push(arguments)
      }
    window.gtag = gtag

    var consent = 'granted'
    gtag('consent', 'default', {
      ad_storage: consent,
      analytics_storage: consent,
      ad_user_data: consent,
      ad_personalization: consent,
    })
    gtag('js', new Date())
    gtag('config', measurementId)

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

    var cached = localStorage.getItem('eea')
    if (cached === '1' || cached === '0') {
      return cached === '1'
    }

    try {
      var response = await fetch('/eea/v1')
      if (!response.ok) throw new Error('bad status')

      var payload = (await response.text()).trim()
      if (payload !== '0' && payload !== '1') throw new Error('invalid payload')
      localStorage.setItem('eea', payload)
      return payload === '1'
    } catch {
      return false
    }
  }

  async function init() {
    var hasKlaro = localStorage.getItem('klaro') !== null
    var isEEA = await getIsEEA(hasKlaro)

    if (hasKlaro || isEEA) {
      loadKlaro()
    }

    canTrack = hasKlaro || !isEEA
    if (canTrack) {
      loadGa()
    }
    trackUrl()
  }

  installNavTracker()
  void init()
})()
