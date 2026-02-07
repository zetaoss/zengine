// mediawiki.ts
import mwapi from '@/utils/mwapi'

export default async function titleExist(title: string): Promise<boolean> {
  const results = await titlesExist([title])
  return results[title] === true
}

interface PageV2 {
  title: string
  missing?: unknown
}

interface NormalizedEntry {
  from: string
  to: string
}

interface DataV2 {
  query?: {
    normalized?: NormalizedEntry[]
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

export async function titlesExist(
  titles: string[],
): Promise<Record<string, boolean>> {
  const uniqueTitles = [...new Set(titles.map(t => t.trim()).filter(Boolean))]
  if (!uniqueTitles.length) return {}

  const chunks = chunkArray(uniqueTitles, 50)
  const results: Record<string, boolean> = {}

  await Promise.all(
    chunks.map(async chunk => {
      const [data, err] = await mwapi.get<DataV2>({
        action: 'query',
        titles: chunk.join('|'),
      })
      if (err) {
        console.error(err)
        return
      }
      const query = data?.query
      const pages = query?.pages ?? []
      const normalizedMap = new Map(
        (query?.normalized ?? []).map(({ from, to }) => [from, to]),
      )
      const pageExistsByTitle = new Map(
        pages.map(page => [
          page.title,
          !Object.prototype.hasOwnProperty.call(page, 'missing'),
        ]),
      )
      const resolveTitle = (title: string) => normalizedMap.get(title) ?? title

      for (const originalTitle of chunk) {
        results[originalTitle] =
          pageExistsByTitle.get(resolveTitle(originalTitle)) === true
      }
    }),
  )

  return results
}
