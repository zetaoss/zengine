export type ZConf = {
  policy: string
  gaMeasurementId: string
  avatarBaseUrl: string
  adClient: string
  adSlots: string[]
}

export function getZConf(): ZConf {
  return (window as Window & { ZCONF: ZConf }).ZCONF
}
