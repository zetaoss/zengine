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

  constructor(
    code: number,
    kind: HttpyErrorKind,
    message: string,
  ) {
    super(message)
    this.name = 'HttpyError'
    this.kind = kind
    this.code = code

    Object.setPrototypeOf(this, new.target.prototype)
  }

  static fromUnknown(e: unknown): HttpyError {
    if (e instanceof HttpyError) return e

    if (e instanceof Error) {
      return new HttpyError(0, 'NETWORK', e.message)
    }

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
  private async parseResponse<TRaw = unknown, TSelected = TRaw>(
    res: Response,
    selector?: (raw: TRaw) => TSelected,
    options: HttpyOptions<TSelected> = {},
  ): Promise<HttpyResult<TSelected>> {
    const text = await res.text()

    if (!res.ok) {
      return [null, new HttpyError(res.status, 'HTTP', `${res.status}`)]
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

  private buildUrl(url: string, params: HttpyParams | undefined): string {
    if (!params || Object.keys(params).length === 0) return url

    const query = new URLSearchParams()
    for (const [k, v] of Object.entries(params)) {
      if (v == null) continue
      query.append(k, String(v))
    }

    return query.toString() ? `${url}?${query.toString()}` : url
  }

  async get<TRaw = unknown, TSelected = TRaw>(
    url: string,
    params: HttpyParams = {},
    selector?: (raw: TRaw) => TSelected,
    options: HttpyOptions<TSelected> = {},
  ): Promise<HttpyResult<TSelected>> {
    try {
      const fullUrl = this.buildUrl(url, params)
      const res = await fetch(fullUrl)
      return this.parseResponse<TRaw, TSelected>(res, selector, options)
    } catch (e) {
      return [null, HttpyError.fromUnknown(e)]
    }
  }

  async delete<T = unknown>(
    url: string,
    options: HttpyOptions<T> = {},
  ): Promise<HttpyResult<T>> {
    try {
      const res = await fetch(url, { method: 'DELETE' })
      return this.parseResponse<T, T>(res, undefined, options)
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
      const res = await fetch(url, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      return this.parseResponse<T, T>(res, undefined, options)
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
      const isForm = body instanceof FormData

      const res = await fetch(url, {
        method: 'POST',
        headers: isForm ? undefined : { 'Content-Type': 'application/json' },
        body: isForm ? body : JSON.stringify(body),
      })

      return this.parseResponse<T, T>(res, undefined, options)
    } catch (e) {
      return [null, HttpyError.fromUnknown(e)]
    }
  }
}

const httpy = new Httpy()
export default httpy
