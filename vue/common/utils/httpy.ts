// @common/utils/httpy.ts

export type HttpyResult<T> = [T, null] | [null, HttpyError]

export interface HttpyParams {
  [key: string]: string | number | boolean | null | undefined
}

export type HttpyErrorKind = 'HTTP' | 'NETWORK' | 'BAD_DATA'

export interface HttpyOptions<T = unknown> {
  validate?: (value: unknown) => value is T
}

export class HttpyError extends Error {
  kind: HttpyErrorKind
  code: number

  constructor(code: number, kind: HttpyErrorKind, message: string) {
    super(message)
    this.name = 'HttpyError'
    this.kind = kind
    this.code = code
    Object.setPrototypeOf(this, new.target.prototype)
  }

  static fromUnknown(e: unknown): HttpyError {
    if (e instanceof HttpyError) return e
    if (e instanceof Error) return new HttpyError(0, 'NETWORK', e.message)
    return new HttpyError(0, 'NETWORK', String(e ?? 'unknown error'))
  }

  override toString(): string {
    return `${this.kind} ${this.message}`
  }

  toJSON() {
    return {
      name: this.name,
      kind: this.kind,
      code: this.code,
      message: this.message,
    }
  }
}

class Httpy {
  private buildUrl(url: string, params?: HttpyParams): string {
    if (!params || Object.keys(params).length === 0) return url

    const query = new URLSearchParams()
    for (const [k, v] of Object.entries(params)) {
      if (v == null) continue
      query.append(k, String(v))
    }

    const qs = query.toString()
    return qs ? `${url}?${qs}` : url
  }

  private async parseResponse<TRaw = unknown, TSelected = TRaw>(
    res: Response,
    selector?: (raw: TRaw) => TSelected,
    options: HttpyOptions<TSelected> = {},
  ): Promise<HttpyResult<TSelected>> {
    const text = await res.text()

    if (!res.ok) {
      // JSON 에러면 message를 최대한 뽑고, 아니면 status만
      let message = `${res.status}`

      try {
        const parsedUnknown: unknown = JSON.parse(text)
        if (parsedUnknown && typeof parsedUnknown === 'object') {
          const obj = parsedUnknown as { message?: unknown }
          if (typeof obj.message === 'string' && obj.message.length > 0) {
            message = obj.message
          }
        }
      } catch {
        // HTML/text 응답이면 무시
      }

      return [null, new HttpyError(res.status, 'HTTP', message)]
    }

    if (res.status === 204 || res.status === 205) {
      return [undefined as unknown as TSelected, null]
    }

    let parsed: TRaw
    try {
      parsed = JSON.parse(text) as TRaw
    } catch {
      return [null, new HttpyError(res.status, 'BAD_DATA', 'not json')]
    }

    let value: unknown = parsed
    if (selector) {
      try {
        value = selector(parsed)
      } catch {
        return [null, new HttpyError(res.status, 'BAD_DATA', 'selector error')]
      }
    }

    const { validate } = options
    if (validate && !validate(value)) {
      return [null, new HttpyError(res.status, 'BAD_DATA', 'invalid')]
    }

    return [value as TSelected, null]
  }

  private async request<TRaw = unknown, TSelected = TRaw>(
    method: string,
    url: string,
    {
      params,
      body,
      selector,
      options,
    }: {
      params?: HttpyParams
      body?: Record<string, unknown> | FormData
      selector?: (raw: TRaw) => TSelected
      options?: HttpyOptions<TSelected>
    } = {},
  ): Promise<HttpyResult<TSelected>> {
    const fullUrl = this.buildUrl(url, params)
    const isForm = body instanceof FormData

    const res = await fetch(fullUrl, {
      method,
      credentials: 'include',
      headers: {
        Accept: 'application/json',
        ...(isForm ? {} : { 'Content-Type': 'application/json' }),
      },
      body: body ? (isForm ? body : JSON.stringify(body)) : undefined,
    })

    return this.parseResponse<TRaw, TSelected>(res, selector, options ?? {})
  }

  async get<TRaw = unknown, TSelected = TRaw>(
    url: string,
    params: HttpyParams = {},
    selector?: (raw: TRaw) => TSelected,
    options: HttpyOptions<TSelected> = {},
  ): Promise<HttpyResult<TSelected>> {
    try {
      return await this.request<TRaw, TSelected>('GET', url, {
        params,
        selector,
        options,
      })
    } catch (e) {
      return [null, HttpyError.fromUnknown(e)]
    }
  }

  async delete<T = unknown>(
    url: string,
    options: HttpyOptions<T> = {},
  ): Promise<HttpyResult<T>> {
    try {
      return await this.request<T, T>('DELETE', url, { options })
    } catch (e) {
      return [null, HttpyError.fromUnknown(e)]
    }
  }

  async put<T = unknown>(
    url: string,
    body: Record<string, unknown> = {},
    options: HttpyOptions<T> = {},
  ): Promise<HttpyResult<T>> {
    try {
      return await this.request<T, T>('PUT', url, { body, options })
    } catch (e) {
      return [null, HttpyError.fromUnknown(e)]
    }
  }

  async post<T = unknown>(
    url: string,
    body: Record<string, unknown> | FormData = {},
    options: HttpyOptions<T> = {},
  ): Promise<HttpyResult<T>> {
    try {
      return await this.request<T, T>('POST', url, { body, options })
    } catch (e) {
      return [null, HttpyError.fromUnknown(e)]
    }
  }
}

const httpy = new Httpy()
export default httpy
