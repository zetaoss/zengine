// util.ts
export type ApiError = Error | null
export type ApiResult<T> = [T | null, ApiError]

export function wrapError(msg: string, err: ApiError): ApiError {
  return newApiError(`${msg}: ${String(err)}`)
}

export function wrapErrorString(msg: string, err: ApiError): string {
  return String(wrapError(msg, err))
}

function newApiError(err: unknown): ApiError {
  if (err instanceof Error) return err
  return new Error(String(err))
}

async function fetchJson(url: string, init?: RequestInit): Promise<[Response, unknown]> {
  const res = await fetch(url, init)
  let body: unknown = null

  try {
    body = await res.json()
  } catch {
  }

  return [res, body]
}

export async function apiGet<T>(url: string, init?: RequestInit): Promise<ApiResult<T>> {
  try {
    const [res, body] = await fetchJson(url, init)

    if (!res.ok) {
      const msg = `${res.status} ${res.statusText}`
      return [null, new Error(msg)]
    }

    return [body as T, null]
  } catch (err: unknown) {
    return [null, newApiError(err)]
  }
}

export async function actionApiGet<T>(url: string, init?: RequestInit): Promise<ApiResult<T>> {
  try {
    const [res, body] = await fetchJson(url, init)

    if (!res.ok) {
      let msg = `${res.status} ${res.statusText}`

      const b = body as { error?: { info?: string } } | null
      if (b?.error?.info) {
        msg = b.error.info
      }

      return [null, new Error(msg)]
    }

    return [body as T, null]
  } catch (err: unknown) {
    return [null, newApiError(err)]
  }
}

export async function restApiGet<T>(url: string, init?: RequestInit): Promise<ApiResult<T>> {
  try {
    const [res, body] = await fetchJson(url, init)

    if (!res.ok) {
      let msg = `${res.status} ${res.statusText}`

      const b = body as { errors?: Array<{ message?: string }> } | null
      const firstMsg = b?.errors?.[0]?.message
      if (firstMsg && firstMsg.trim() !== '') {
        msg = firstMsg
      }

      return [null, new Error(msg)]
    }

    return [body as T, null]
  } catch (err: unknown) {
    return [null, newApiError(err)]
  }
}

export async function laravelApiGet<T>(url: string, init?: RequestInit): Promise<ApiResult<T>> {
  try {
    const [res, body] = await fetchJson(url, init)

    if (!res.ok) {
      let msg = `${res.status} ${res.statusText}`

      const b = body as {
        message?: string
      } | null

      if (b?.message && b.message.trim() !== '') {
        msg = b.message
      }
      return [null, new Error(msg)]
    }

    return [body as T, null]
  } catch (err: unknown) {
    return [null, newApiError(err)]
  }
}
