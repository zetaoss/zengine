// wapi.ts

export type WapiResult<T> = [T | null, Error | null]

type QueryParams = Record<string, string | number | boolean | null | undefined>

function buildQuery(params: QueryParams): string {
  const usp = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value === null || value === undefined) continue
    usp.append(key, String(value))
  }
  return usp.toString()
}

async function request(url: string, params: QueryParams): Promise<unknown> {
  const qs = buildQuery(params)
  const res = await fetch(`${url}?${qs}`, {
    credentials: 'include',
  })

  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`)
  }

  return await res.json()
}

interface MwUserQueryResponse {
  query?: {
    users?: Array<{
      userid?: number
    }>
  }
}

export async function getUserId(name: string): Promise<WapiResult<number>> {
  try {
    const raw = await request('/w/api.php', {
      action: 'query',
      list: 'users',
      ususers: name,
      format: 'json',
    })

    const data = raw as MwUserQueryResponse
    const uid = data.query?.users?.[0]?.userid
    if (!uid) {
      return [null, new Error('User not found')]
    }

    return [uid, null]
  } catch (e: unknown) {
    const err = e instanceof Error ? e : new Error(String(e))
    return [null, err]
  }
}
