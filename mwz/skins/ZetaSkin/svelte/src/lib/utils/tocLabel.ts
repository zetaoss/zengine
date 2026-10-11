// Same steps as StaticMarkup::tocNode so the server-rendered TOC and this one show the same text.
export default function tocLabel(line: string): string {
  const stripped = line.replace(/<\/?[^>]+>/gi, ' ').trim()
  const ta = document.createElement('textarea')
  ta.innerHTML = stripped
  return ta.value
}
