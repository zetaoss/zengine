// mediawiki.ts
import httpy from '@common/utils/httpy'

interface Page {
  pageid?: number
  ns?: number
  title?: string
  missing?: string
}

interface Data {
  query?: {
    pages?: Record<string, Page>
  }
}

export default async function titleExist(title: string): Promise<boolean> {
  const [data, err] = await httpy.get<Data>('/w/api.php', {
    action: 'query',
    format: 'json',
    titles: title,
  })
  if (err) {
    console.error(err)
    return false
  }
  const pages = data?.query?.pages
  if (!pages) return false
  return Object.keys(pages).some(key => key !== '-1')
}
