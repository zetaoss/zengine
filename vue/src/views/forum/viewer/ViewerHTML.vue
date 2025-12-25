<!-- BoxHTML.vue -->
<script setup lang="ts">
import httpy from '@common/utils/httpy'
import { nextTick, ref, watch } from 'vue'

import linkify from '@/utils/linkify'

import { applyHljs } from './hljs'
import { renderPlainTextWithFences } from './renderText'

interface PreviewData {
  type: 'image' | 'text'
  image?: string
  title?: string
  description?: string
}

const props = defineProps({
  body: { type: String, required: true },
  mode: { type: String as () => 'html' | 'text', default: 'html' },
  previews: { type: Boolean, default: true },
  fencedCode: { type: Boolean, default: false },
})

const rootEl = ref<HTMLElement | null>(null)
const html = ref('')

function escapeHtml(s: string) {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
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

async function getPreviews(concurrency = 3) {
  const matches =
    html.value.match(/<a href="[^"]+"[^>]*external[^>]*>[^<]+<\/a>/g) ?? []

  let index = 0

  async function worker() {
    while (index < matches.length) {
      const match = matches[index++]
      const m = match.match(/"([^"]+)"/)
      if (!m) continue

      const url = m[1]
      const [data] = await httpy.get<PreviewData>('/api/preview', { url })
      if (!data) continue

      html.value = html.value.replace(
        match,
        match.replace('>', ' preview>') + data2html(url, data),
      )
    }
  }

  await Promise.all(Array.from({ length: concurrency }, worker))
}

async function afterRender() {
  await nextTick()
  if (rootEl.value) applyHljs(rootEl.value)
}

watch(
  () => [props.body, props.mode, props.previews, props.fencedCode] as const,
  async () => {
    if (props.mode === 'text') {
      const base = props.fencedCode ? renderPlainTextWithFences(props.body) : renderPlainTextWithFences(props.body)
      html.value = await linkify(base)
      await afterRender()
      return
    }

    html.value = await linkify(props.body)
    if (props.previews) await getPreviews()
    await afterRender()
  },
  { immediate: true },
)

watch(html, afterRender)
</script>

<template>
  <div ref="rootEl" class="ProseMirror" v-html="html" />
</template>

<style lang="scss">
.preview {
  @apply flex border rounded p-3;

  a {
    @apply px-3 text-sm text-gray-700 dark:text-gray-300;

    &:hover {
      @apply no-underline;
    }
  }

  span {
    @apply flex-none w-[20vw] bg-cover bg-center;
  }

  p {
    @apply py-2;
  }

  img {
    @apply inline border-0 m-0;
  }
}
</style>
