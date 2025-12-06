// util.ts
export function pad(n: number): string {
  return String(n).padStart(2, '0')
}

export function toDate(d: Date): string {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}
