// mwapi.ts
import httpy, { HttpyError, type HttpyParams, type HttpyResult } from '@common/utils/httpy'

type MwApiParams = HttpyParams

interface MwApiError {
  code?: string
  info?: string
}

interface MwApiEnvelope {
  error?: MwApiError
}

const DEFAULT_PARAMS: MwApiParams = {
  format: 'json',
  formatversion: '2',
}

function mergeParams(params?: MwApiParams): MwApiParams {
  return {
    ...DEFAULT_PARAMS,
    ...(params ?? {}),
  }
}

function wrapError(error: MwApiError | undefined): HttpyError | null {
  if (!error) return null
  const code = error.code ? ` ${error.code}` : ''
  const info = error.info ? ` ${error.info}` : ''
  const message = `mw error:${code}${info}`.trim()
  return new HttpyError(0, 'BAD_DATA', message)
}

async function get<T>(params: MwApiParams): Promise<HttpyResult<T>> {
  const [data, err] = await httpy.get<MwApiEnvelope & T>('/w/api.php', mergeParams(params))
  if (err) return [null, err]
  if (!data) return [null, new HttpyError(0, 'BAD_DATA', 'empty response')]

  const mwError = wrapError(data.error)
  if (mwError) return [null, mwError]

  return [data as T, null]
}

const mwapi = {
  get,
}

export type { MwApiParams }
export default mwapi
