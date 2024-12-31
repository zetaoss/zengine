import http from '@/utils/http'

async function getLogintoken() {
  const resp: any = await http.get('/w/api.php', {
    params: {
      format: 'json',
      action: 'query',
      meta: 'tokens',
      type: 'login',
    },
  })
  return resp.data.query.tokens.logintoken
}

export default async function doLogin(username: string, password: string, loginreturnurl: string) {
  const logintoken = await getLogintoken()
  const params = {
    format: 'json',
    action: 'clientlogin',
    loginreturnurl,
    logintoken,
    username,
    password,
    rememberMe: '1',
  }
  const headers = { 'Content-Type': 'multipart/form-data' }
  const resp: any = await http.post('/w/api.php', params, { headers })
  return resp.data.clientlogin
}
