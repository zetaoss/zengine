export function maskEmail(email: string): string {
  const trimmed = email.trim()
  const at = trimmed.indexOf('@')
  if (at <= 0 || at === trimmed.length - 1) return ''

  const local = trimmed.slice(0, at)
  const domain = trimmed.slice(at + 1)

  const maskedLocal =
    local.length <= 2 ? `${local[0] ?? ''}*` : `${local[0] ?? ''}${'*'.repeat(local.length - 2)}${local[local.length - 1] ?? ''}`

  const dot = domain.lastIndexOf('.')
  if (dot <= 1 || dot === domain.length - 1) {
    return `${maskedLocal}@${domain[0] ?? ''}***`
  }

  const host = domain.slice(0, dot)
  const tld = domain.slice(dot)
  const maskedHost = host.length <= 2 ? `${host[0] ?? ''}*` : `${host[0] ?? ''}${'*'.repeat(host.length - 2)}${host[host.length - 1] ?? ''}`

  return `${maskedLocal}@${maskedHost}${tld}`
}
