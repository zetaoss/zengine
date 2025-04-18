<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import http from '@/utils/http'
import '../tiptap/ProseMirror.scss'
import linkify from '@/utils/linkify'

interface PreviewData {
  type: 'image' | 'text'
  image?: string
  title?: string
  description?: string
}

const props = defineProps({
  body: { required: true, type: String },
})

const html = ref('');
const renderComponent = ref(1);

function imageToHTML(url: string) {
  return `<div><img class="border max-w-[50vw] max-h-[70vh]" src="${url}" /></div>`
}

function textToHTML(url: string, data: PreviewData): string {
  const { host } = new URL(url);
  return `<div class="btn preview"><span style="background-image: url('${data.image}')"></span><a href='${url}'><b>${data.title}</b><p>${data.description}</p><img src="//www.google.com/s2/favicons?domain=${host}"> ${host}</span></a></div>`
}

function data2html(url: string, data: PreviewData): string {
  if (data.type === 'image') return imageToHTML(url);
  if (data.type === 'text') return textToHTML(url, data);
  return '';
}

async function previewMatch(match: string) {
  const m = match.match(/"([^"]+)"/);
  if (!m) return;
  const url = m[1];
  const resp = await http.get('/api/preview', { params: { url } });
  html.value = html.value.replace(match, match.replace('>', ' preview>') + data2html(url, resp.data));
}

async function getPreviews() {
  const matches = html.value.match(/<a href="[^"]+"[^>]*external[^>]*>[^<]+<\/a>/g);
  if (!matches) return;
  for (const match of matches) {
    await previewMatch(match)
  }
}

onMounted(async () => {
  html.value = await linkify(props.body);
  await getPreviews();
})

watch(() => html.value, async () => {
  await nextTick();
  renderComponent.value++;
})
</script>

<template>
  <div class="ProseMirror" v-html="html" />
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
