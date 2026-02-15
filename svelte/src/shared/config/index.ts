// $shared/config/index.ts
type RuntimeConfig = {
  avatarBaseUrl?: string
}

let cached: string | null = null
let checked = false

export function getAvatarBaseUrl(): string {
  if (checked) {
    return cached ?? ''
  }

  checked = true

  const win = window as unknown as {
    __CONFIG__?: RuntimeConfig
  }

  const raw = win.__CONFIG__?.avatarBaseUrl

  if (typeof raw !== 'string' || raw.length === 0) {
    console.error('[config] missing runtime config: avatarBaseUrl')
    cached = null
    return ''
  }

  cached = raw
  return cached
}
