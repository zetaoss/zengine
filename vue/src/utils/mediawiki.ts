import http from '@/utils/http'

export default async function titleExist(title: string) {
  const resp: any = await http.get('/w/api.php', { params: { action: 'query', format: 'json', titles: title } })
  return Object.prototype.hasOwnProperty.call(resp.data.query.pages, -1)
}
