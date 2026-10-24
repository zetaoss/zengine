export function getAge(timestamp: string, now = Date.now()): string {
  if (!timestamp) return ''
  const d = new Date(timestamp)
  if (isNaN(d.getTime())) return ''

  const diffSec = Math.floor((now - d.getTime()) / 1000)

  if (diffSec < 0) return '방금 전'
  if (diffSec < 60) return `${diffSec}초 전`

  const min = Math.floor(diffSec / 60)
  if (min < 60) return `${min}분 전`

  const hour = Math.floor(min / 60)
  if (hour < 24) return `${hour}시간 전`

  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

export function formatDateTime(value: string | null | undefined): string {
  if (!value) return ''
  const d = new Date(value)
  return isNaN(d.getTime()) ? '' : d.toLocaleString()
}
