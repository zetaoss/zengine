export function maskEmail(email: string): string {
  const [name = '', domainPart = ''] = email.split('@')
  const [host = '', tld = ''] = domainPart.split('.')

  const maskPart = (part: string): string => part.slice(0, 2) + '*'.repeat(Math.max(0, part.length - 2))

  return `${maskPart(name)}@${maskPart(host)}.${maskPart(tld)}`
}
