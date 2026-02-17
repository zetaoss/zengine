export function maskEmail(email: string): string {
  return email.replace(/[^@.]+/g, (part) => part.slice(0, 2) + '*'.repeat(Math.max(0, part.length - 2)))
}
