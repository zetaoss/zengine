async function request(method: string, url: string, body?: object) {
  const init: RequestInit = {
    method,
    headers: { 'Content-Type': 'application/json' },
    ...(body && { body: JSON.stringify(body) })
  }
  const response = await fetch(url, init)
  return response.json()
}

const http = {
  get: (url: string) => request('GET', url),
  post: (url: string, data: object) => request('POST', url, data),
  put: (url: string, data: object) => request('PUT', url, data),
  delete: (url: string, data?: object) => request('DELETE', url, data),
}

export default http
