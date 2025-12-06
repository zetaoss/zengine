// action.ts
import { actionApiGet, wrapError, type ApiResult } from './util'

export interface Contribution {
  timestamp: string
  title: string
  revid: number
}

interface UserContribsResponse {
  query?: {
    usercontribs?: Contribution[]
  }
}

export async function getContributions(userId: number): Promise<ApiResult<Contribution[]>> {
  const params = new URLSearchParams({
    action: 'query',
    format: 'json',
    list: 'usercontribs',
    ucuserids: String(userId),
    ucprop: 'ids|title|timestamp',
    uclimit: '10',
  })
  const [data, err] = await actionApiGet<UserContribsResponse>(`/w/api.php?${params}`)
  if (err) return [null, wrapError('UserContribsResponse err', err)]
  return [data?.query?.usercontribs ?? [], null]
}
