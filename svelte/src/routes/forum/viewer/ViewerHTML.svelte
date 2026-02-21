<script lang="ts">
  import { tick } from 'svelte'

  import httpy from '$shared/utils/httpy'
  import linkify from '$shared/utils/linkify'

  import { applyHljs } from './hljs'
  import { renderPlainTextWithFences } from './renderText'

  interface PreviewData {
    type: 'image' | 'text'
    image?: string
    title?: string
    description?: string
  }

  export let body: string
  export let mode: 'html' | 'text' = 'html'
  export let previews = true
  export let fencedCode = false

  let rootEl: HTMLElement | null = null
  let html = ''
  let token = 0

  function escapeHtml(s: string) {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#039;')
  }

  function imageToHTML(url: string) {
    return `<div><img class="border max-w-[50vw] max-h-[70vh]" src="${url}" /></div>`
  }

  function textToHTML(url: string, data: PreviewData) {
    const { host } = new URL(url)
    const title = escapeHtml(data.title ?? url)
    const desc = escapeHtml(data.description ?? '')
    const image = data.image ?? ''

    return `
      <div class="btn preview">
        <span style="background-image: url('${image}')"></span>
        <a href="${url}">
          <b>${title}</b>
          <p>${desc}</p>
          <img src="//www.google.com/s2/favicons?domain=${host}"> ${host}
        </a>
      </div>
    `.trim()
  }

  function data2html(url: string, data: PreviewData) {
    if (data.type === 'image') return imageToHTML(url)
    if (data.type === 'text') return textToHTML(url, data)
    return ''
  }

  async function getPreviews() {
    const matches = html.match(/<a href="[^"]+"[^>]*external[^>]*>[^<]+<\/a>/g) ?? []

    let index = 0

    async function worker() {
      while (index < matches.length) {
        const match = matches[index++]
        if (!match) continue
        const m = match.match(/"([^"]+)"/)
        if (!m) continue

        const url = m[1]
        if (!url) continue
        const [data] = await httpy.get<PreviewData>('/api/preview', { url })
        if (!data) continue

        html = html.replace(match, match.replace('>', ' preview>') + data2html(url, data))
      }
    }

    await Promise.all([worker(), worker(), worker()])
  }

  async function afterRender() {
    await tick()
    if (rootEl) applyHljs(rootEl)
  }

  async function compute() {
    const current = ++token
    if (mode === 'text') {
      const base = fencedCode ? renderPlainTextWithFences(body) : renderPlainTextWithFences(body)
      const linked = (await linkify([base]))[0] ?? ''
      if (current !== token) return
      html = linked
      await afterRender()
      return
    }

    const linked = (await linkify([body]))[0] ?? ''
    if (current !== token) return
    html = linked
    if (previews) await getPreviews()
    await afterRender()
  }

  $: if (body !== undefined && mode !== undefined && previews !== undefined && fencedCode !== undefined) {
    void compute()
  }
</script>

<!-- eslint-disable-next-line svelte/no-at-html-tags -->
<div bind:this={rootEl} class="ProseMirror">{@html html}</div>

<style lang="postcss">
  @reference 'tailwindcss';

  :global(.preview) {
    @apply flex border rounded p-3;
  }

  :global(.preview a) {
    @apply px-3 text-sm text-gray-700 dark:text-gray-300;
  }

  :global(.preview a:hover) {
    @apply no-underline;
  }

  :global(.preview span) {
    @apply flex-none w-[20vw] bg-cover bg-center;
  }

  :global(.preview p) {
    @apply py-2;
  }

  :global(.preview img) {
    @apply inline border-0 m-0;
  }
</style>
