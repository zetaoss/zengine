async function httpDo(method: string, url: string, data: Object) {
  const headers = { 'Content-Type': 'application/json' }
  const init = method === 'GET' ? { method, headers } : { method, headers, body: JSON.stringify(data) }
  const response = await fetch(url, init)
  return response.json()
}

const http = {
  get: (url = '', data = {}) => httpDo('GET', url, data),
  post: (url = '', data = {}) => httpDo('POST', url, data),
  put: (url = '', data = {}) => httpDo('PUT', url, data),
  delete: (url = '', data = {}) => httpDo('DELETE', url, data),
}

export default http
