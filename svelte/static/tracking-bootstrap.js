;(function () {
  function readCookie(name) {
    var key = name + '='
    var tokens = document.cookie ? document.cookie.split(';') : []
    for (var i = 0; i < tokens.length; i += 1) {
      var token = tokens[i].trim()
      if (!token.startsWith(key)) continue
      var raw = token.slice(key.length)
      try {
        return decodeURIComponent(raw)
      } catch {
        return raw
      }
    }
    return ''
  }

  function canBootGtag(zconf) {
    var policy = zconf && zconf.policy
    var rawConsent = readCookie('consent')
    var allowed = policy === 'standard'
    if (!rawConsent) return allowed

    var params = new URLSearchParams(rawConsent)
    var analyticsGranted = params.get('analytics_storage') === 'granted'
    var adsGranted = params.get('ad_storage') === 'granted'
    return analyticsGranted || adsGranted
  }

  function preloadGtag() {
    var w = window
    var measurementId = w.ZCONF && w.ZCONF.gaMeasurementId
    if (!measurementId || typeof measurementId !== 'string') return
    if (!canBootGtag(w.ZCONF)) return

    var src = 'https://www.googletagmanager.com/gtag/js?id=' + encodeURIComponent(measurementId)
    if (document.querySelector('script[src="' + src + '"]')) return

    var script = document.createElement('script')
    script.async = true
    script.src = src
    document.head.appendChild(script)
  }

  window.zetaTrackingBootstrap = {
    preloadGtag: preloadGtag,
  }
})()
