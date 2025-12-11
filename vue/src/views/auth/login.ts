// login.ts
import httpy, { type HttpyError } from '@common/utils/httpy'

interface LoginTokenData {
  query: {
    tokens: {
      logintoken: string
    }
  }
}

interface ClientLogin {
  status: string
  message?: string
  username?: string
  redirecturl?: string
}

interface ClientLoginData {
  clientlogin: ClientLogin
}

async function getLogintoken(): Promise<[string, HttpyError | null]> {
  const [data, err] = await httpy.get<LoginTokenData>('/w/api.php', {
    format: 'json',
    action: 'query',
    meta: 'tokens',
    type: 'login',
  })

  if (err) return ['', err]

  return [data.query.tokens.logintoken, null]
}

export default async function doLogin(
  username: string,
  password: string,
  loginreturnurl: string,
): Promise<[ClientLogin | null, HttpyError | null]> {
  const [logintoken, err] = await getLogintoken()
  if (err) return [null, err]

  const params = {
    format: 'json',
    action: 'clientlogin',
    loginreturnurl,
    logintoken,
    username,
    password,
    rememberMe: '1',
  }

  const form = new FormData()
  Object.entries(params).forEach(([k, v]) => {
    if (v != null) form.append(k, String(v))
  })

  const [data, err2] = await httpy.post<ClientLoginData>('/w/api.php', form)
  if (err2) return [null, err2]

  return [data.clientlogin, null]
}
