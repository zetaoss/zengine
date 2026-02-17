export function maskEmail(email: string): string {
  return Array.from(email, (ch, i) => {
    if (ch === '@') return '@'
    return i % 2 === 0 ? ch : '*'
  }).join('')
}
