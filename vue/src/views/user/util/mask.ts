// mask.ts

export function maskEmail(email: string): string {
  return email
    .split(/([@.])/)
    .map(p => {
      if (p === '@' || p === '.') return p
      if (p.length <= 2) return p
      return p[0] + '*'.repeat(p.length - 2) + p[p.length - 1]
    })
    .join('')
}
