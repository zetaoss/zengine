import getRLCONF from './rlconf'

function decodeTitle(value: string) {
  return value.replaceAll('_', ' ').trim()
}

export default function getCurrentTitle() {
  const { wgPageName, wgTitle } = getRLCONF()
  if (wgPageName) return decodeTitle(wgPageName)
  if (wgTitle) return decodeTitle(wgTitle)

  const heading = document.getElementById('firstHeading')?.textContent?.trim()
  if (heading && heading !== 'index.php') return heading

  const url = new URL(window.location.href)
  const titleParam = url.searchParams.get('title')
  if (titleParam) return decodeTitle(decodeURIComponent(titleParam))

  const pathSegments = decodeURIComponent(url.pathname).split('/').filter(Boolean)
  const lastSegment = pathSegments[pathSegments.length - 1]
  if (lastSegment && lastSegment !== 'index.php') return decodeTitle(lastSegment)

  return '현재'
}
