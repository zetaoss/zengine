// mediawiki.ts
import mwapi from '@/utils/mwapi'

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
  const results = await titlesExist([title])
  return results[title] === true
}

interface PageV2 {
  title: string
  missing?: unknown
}

interface DataV2 {
  query?: {
    pages?: PageV2[]
  }
}

function chunkArray<T>(items: T[], size: number): T[][] {
  const chunks: T[][] = []
  for (let i = 0; i < items.length; i += size) {
    chunks.push(items.slice(i, i + size))
  }
  return chunks
}

export async function titlesExist(titles: string[]): Promise<Record<string, boolean>> {
  const uniqueTitles = Array.from(new Set(titles.map(t => t.trim()).filter(Boolean)))
  if (uniqueTitles.length === 0) return {}

  const chunks = chunkArray(uniqueTitles, 50)
  const results: Record<string, boolean> = {}

  await Promise.all(chunks.map(async (chunk) => {
    const [data, err] = await mwapi.get<DataV2>({
      action: 'query',
      titles: chunk.join('|'),
    })
    if (err) {
      console.error(err)
      return
    }
    const pages = data?.query?.pages ?? []
    for (const page of pages) {
      results[page.title] = !Object.prototype.hasOwnProperty.call(page, 'missing')
    }
  }))

  return results
}
