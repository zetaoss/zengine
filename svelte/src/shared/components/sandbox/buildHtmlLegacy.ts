import buildHtml from './buildHtml'

// buildHtmlLegacy.ts
export default function buildHtmlLegacy(id: string, html: string, js: string): string {
  return buildHtml(id, '', html, js)
}
