import httpy from '$shared/utils/httpy'

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
    normalized?: NormalizedEntry[] | Record<string, NormalizedEntry>
    pages?: PageV2[] | Record<string, PageV2>
  }
  error?: {
    code?: string
    info?: string
  }
}

function chunkArray<T>(items: T[], size: number): T[][] {
  const chunks: T[][] = []
  for (let i = 0; i < items.length; i += size) {
    chunks.push(items.slice(i, i + size))
  }
  return chunks
}

function asArray<T>(value: T[] | Record<string, T> | undefined): T[] {
  if (!value) return []
  return Array.isArray(value) ? value : Object.values(value)
}

export async function titlesExist(titles: string[]): Promise<Record<string, boolean>> {
  const normalizedInputs = titles
    .map((rawTitle) => ({ rawTitle, normalizedTitle: rawTitle.trim() }))
    .filter(({ normalizedTitle }) => normalizedTitle.length > 0)
  const uniqueTitles = [...new Set(normalizedInputs.map(({ normalizedTitle }) => normalizedTitle))]
  if (!uniqueTitles.length) return {}

  const chunks = chunkArray(uniqueTitles, 50)
  const resultByNormalizedTitle: Record<string, boolean> = {}

  await Promise.all(
    chunks.map(async (chunk) => {
      const [data, err] = await httpy.get<DataV2>('/w/api.php', {
        action: 'query',
        format: 'json',
        formatversion: '2',
        titles: chunk.join('|'),
      })
      if (err) {
        console.error(err)
        return
      }

      const query = data?.query
      const pages = asArray(query?.pages)
      const normalizedMap = new Map(asArray(query?.normalized).map(({ from, to }) => [from, to]))
      const pageExistsByTitle = new Map(pages.map((page) => [page.title, !Object.prototype.hasOwnProperty.call(page, 'missing')]))
      const resolveTitle = (title: string) => normalizedMap.get(title) ?? title

      for (const originalTitle of chunk) {
        resultByNormalizedTitle[originalTitle] = pageExistsByTitle.get(resolveTitle(originalTitle)) === true
      }
    }),
  )

  const results: Record<string, boolean> = {}
  for (const { rawTitle, normalizedTitle } of normalizedInputs) {
    const exists = resultByNormalizedTitle[normalizedTitle]
    if (exists !== undefined) {
      results[rawTitle] = exists
    }
  }

  return results
}
