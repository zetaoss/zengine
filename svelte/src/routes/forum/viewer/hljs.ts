import hljs from 'highlight.js/lib/core'
import c from 'highlight.js/lib/languages/c'
import cpp from 'highlight.js/lib/languages/cpp'
import go from 'highlight.js/lib/languages/go'
import php from 'highlight.js/lib/languages/php'
import xml from 'highlight.js/lib/languages/xml'

hljs.registerLanguage('c', c)
hljs.registerLanguage('cpp', cpp)
hljs.registerLanguage('go', go)
hljs.registerLanguage('php', php)
hljs.registerLanguage('xml', xml)
hljs.registerLanguage('html', xml)

export function applyHljs(root: HTMLElement) {
  root.querySelectorAll('pre code').forEach((el) => {
    const codeEl = el as HTMLElement
    if (codeEl.classList.contains('hljs')) return
    hljs.highlightElement(codeEl)
  })
}
