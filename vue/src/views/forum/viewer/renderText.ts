// renderText.ts

function escapeHtml(s: string) {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
}

function escapeCode(s: string) {
  return escapeHtml(s).replace(/\r\n/g, '\n')
}

const FENCE_RE = /```([a-zA-Z0-9_+-]+)?\n([\s\S]*?)```/g

export function renderFencedCode(text: string) {
  return text.replace(
    FENCE_RE,
    (_, lang = 'plaintext', code) =>
      `<pre><code class="language-${lang}">${escapeCode(code)}</code></pre>`,
  )
}

export function renderPlainTextWithFences(text: string) {
  const parts: string[] = []
  let lastIndex = 0

  text.replace(FENCE_RE, (match, lang = 'plaintext', code, offset) => {
    const before = text.slice(lastIndex, offset)
    parts.push(escapeHtml(before).replace(/\n/g, '<br>'))

    parts.push(
      `<pre><code class="language-${lang}">${escapeCode(code)}</code></pre>`,
    )

    lastIndex = offset + match.length
    return match
  })

  const tail = text.slice(lastIndex)
  parts.push(escapeHtml(tail).replace(/\n/g, '<br>'))

  return parts.join('')
}
