export type ZConf = {
  policy: string
  gaMeasurementId: string
  avatarBaseUrl: string
  adSense: {
    client: string
    slots: string[]
  }
}

type RawZConf = Partial<{
  policy: string
  gaMeasurementId: string
  avatarBaseUrl: string
  adSense: {
    client?: string
    slots?: string[]
    slotTop?: string
    slotBottom?: string
  }
}>

function normalizeZConf(raw: RawZConf | undefined): ZConf {
  const adSense = raw?.adSense ?? {}
  const slots = Array.isArray(adSense.slots) && adSense.slots.length > 0 ? adSense.slots : [adSense.slotTop ?? '', adSense.slotBottom ?? '']

  return {
    policy: raw?.policy ?? 'strict',
    gaMeasurementId: raw?.gaMeasurementId ?? '',
    avatarBaseUrl: raw?.avatarBaseUrl ?? '',
    adSense: {
      client: adSense.client ?? '',
      slots,
    },
  }
}

export function getZConf(): ZConf {
  const win = window as Window & { ZCONF?: RawZConf }
  return normalizeZConf(win.ZCONF)
}
