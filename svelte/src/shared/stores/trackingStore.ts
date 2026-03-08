import { readCookie } from '$shared/utils/cookie'
import { getZConf } from '$shared/utils/zConf'

type Policy = 'strict' | 'standard'
type ConsentValue = 'granted' | 'denied'

type ConsentGrants = {
  analyticsStorage: ConsentValue
  adStorage: ConsentValue
  adUserData: ConsentValue
  adPersonalization: ConsentValue
}

export type TrackingState = {
  consentAnalytics: boolean
  consentAds: boolean
  grants: ConsentGrants
  hasConsentCookie: boolean
  policy: Policy
  canBootGtag: boolean
  canShowAds: boolean
}

let trackingGate: () => boolean = () => true

function parseValue(raw: string | null): ConsentValue {
  return raw === 'granted' ? 'granted' : 'denied'
}

function getConsentGrants() {
  const rawConsent = readCookie('consent')
  const policy = getZConf().policy as Policy
  const hasConsentCookie = rawConsent.length > 0

  if (!rawConsent) {
    const granted: ConsentValue = policy === 'standard' ? 'granted' : 'denied'
    return {
      grants: {
        analyticsStorage: granted,
        adStorage: granted,
        adUserData: granted,
        adPersonalization: granted,
      },
      hasConsentCookie,
      policy,
    }
  }

  const params = new URLSearchParams(rawConsent)
  return {
    grants: {
      analyticsStorage: parseValue(params.get('analytics_storage')),
      adStorage: parseValue(params.get('ad_storage')),
      adUserData: parseValue(params.get('ad_user_data')),
      adPersonalization: parseValue(params.get('ad_personalization')),
    },
    hasConsentCookie,
    policy,
  }
}

export function setTrackingGate(evaluator: (() => boolean) | undefined): void {
  trackingGate = evaluator || (() => true)
}

export function getTrackingState(): TrackingState {
  const { grants, hasConsentCookie, policy } = getConsentGrants()
  const consentAnalytics = grants.analyticsStorage === 'granted'
  const consentAds = grants.adStorage === 'granted'

  return {
    consentAnalytics,
    consentAds,
    grants,
    hasConsentCookie,
    policy,
    canBootGtag: consentAnalytics || consentAds,
    canShowAds: trackingGate() && consentAds,
  }
}

export function pushAds(): boolean {
  try {
    window.adsbygoogle = window.adsbygoogle || []
    window.adsbygoogle.push({})
    return true
  } catch {
    return false
  }
}
