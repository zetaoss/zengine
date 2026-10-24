// Shared by the definition and evaluation tabs, whose loops (doc-define-loop, doc-eval-loop) have the same
// settings and run status.
import httpy from '$shared/utils/httpy'

export interface LastRun {
  at: string
  ok: boolean
  page_id?: number
  page_title?: string
  model?: string
  eval_id?: number
  error?: string
}

export interface QueuedTask {
  id: string
  state: 'active' | 'pending' | 'scheduled' | 'failed'
  page_id?: number
  page_title?: string
  next_process_at?: string
  failed_at?: string
  last_error?: string
}

// One loop's settings: interval, model, strategy.
export type LoopSettings = { interval: string; model: string; strategy: string }

export const intervalOptions = [
  { value: '0s', label: '끔' },
  { value: '10m', label: '10분' },
  { value: '30m', label: '30분' },
  { value: '1h', label: '1시간' },
  { value: '2h', label: '2시간' },
  { value: '6h', label: '6시간' },
  { value: '12h', label: '12시간' },
  { value: '24h', label: '24시간' },
]

// The interval options, plus value itself when it is not one of them (set in the config file).
export function intervalOptionsWith(value: string) {
  return intervalOptions.some((o) => o.value === value) ? intervalOptions : [...intervalOptions, { value, label: value }]
}

export const strategyOptions = [
  { value: 'random', label: '랜덤' },
  { value: 'oldest', label: '오랜순' },
  { value: 'newest', label: '최신순' },
]

export const strategyLabels: Record<string, string> = Object.fromEntries(strategyOptions.map((o) => [o.value, o.label]))

export const queueStateLabels: Record<QueuedTask['state'], string> = {
  active: '수행 중',
  pending: '대기',
  scheduled: '예약',
  failed: '실패',
}

export function formatModelLabel(id: string): string {
  const i = id.indexOf('/')
  if (i === -1) return id
  return `${id.slice(0, i)} | ${id.slice(i + 1)}`
}

// yyyy-MM-dd HH:mm in local time, the hover text of the date columns.
export function formatDateTime(iso: string): string {
  if (!iso) return ''
  const d = new Date(iso)
  if (isNaN(d.getTime())) return iso
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function formatTime(iso: string): string {
  const d = new Date(iso)
  if (isNaN(d.getTime())) return iso
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// Time left until iso as m:ss or h:mm:ss.
export function formatCountdown(iso: string, now: number): string {
  const totalSecs = Math.max(0, Math.round((new Date(iso).getTime() - now) / 1000))
  const h = Math.floor(totalSecs / 3600)
  const m = Math.floor((totalSecs % 3600) / 60)
  const s = totalSecs % 60
  const pad = (n: number) => String(n).padStart(2, '0')
  if (h > 0) {
    return `${h}:${pad(m)}:${pad(s)}`
  }
  return `${pad(m)}:${pad(s)}`
}

// Configured bob models in fallback order, for the model dropdowns.
export async function fetchModelIds(): Promise<string[]> {
  const [data, err] = await httpy.get<{
    usage: { data: Array<{ id: string }> } | null
    models: { data: Array<{ id: string; enabled: boolean }> }
  }>('/api/ai/models')
  if (err || !data) return []
  return data.usage?.data.map((m) => m.id) ?? data.models.data.filter((m) => m.enabled).map((m) => m.id)
}

// The wikitext of a revision, from the REST API or else the action API; an Error when neither returns it.
export async function fetchRevisionSource(revid: number): Promise<string | Error> {
  try {
    const resp = await fetch(`/w/rest.php/v1/revision/${revid}`)
    if (resp.ok) {
      const data = await resp.json()
      if (typeof data.source === 'string') return data.source
    }
    const queryResp = await fetch(`/w/api.php?action=query&prop=revisions&revids=${revid}&rvprop=content&rvslots=main&format=json`)
    if (queryResp.ok) {
      const queryData = await queryResp.json()
      const pages = queryData.query?.pages
      if (pages) {
        const page = Object.values(pages)[0] as { revisions?: Array<{ slots?: { main?: { '*'?: string; content?: string } } }> }
        const src = page?.revisions?.[0]?.slots?.main?.['*'] ?? page?.revisions?.[0]?.slots?.main?.content
        if (typeof src === 'string') return src
      }
    }
    return new Error('리비전 내용을 불러오지 못했습니다.')
  } catch (e: unknown) {
    return new Error(e instanceof Error ? e.message : '네트워크 오류가 발생했습니다.')
  }
}

// text split at maxBytes of UTF-8: the part a step read and the rest.
export function splitAtBytes(text: string, maxBytes: number): { read: string; remaining: string } {
  const bytes = new TextEncoder().encode(text)
  if (bytes.length <= maxBytes) {
    return { read: text, remaining: '' }
  }
  const decoder = new TextDecoder('utf-8')
  return { read: decoder.decode(bytes.slice(0, maxBytes)), remaining: decoder.decode(bytes.slice(maxBytes)) }
}
