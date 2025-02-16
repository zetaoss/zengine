<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue';
import TheSplit from '@/components/TheSplit.vue';

const htmlCode = ref('<h1>Hello, World!</h1>');
const jsCode = ref('console.log(new Date());');
const iframeRef = ref<HTMLIFrameElement | null>(null);
const logs = ref<{ type: string; message: unknown[] }[]>([]);
const consoleContainer = ref<HTMLElement | null>(null);

const generateIframeContent = () => {
  const sanitizedHtml = htmlCode.value.trim();
  const wrappedHtml = sanitizedHtml.startsWith('<html>') ? sanitizedHtml : `<html><body>${sanitizedHtml}</body></html>`;

  return `
    ${wrappedHtml}
    <script>
      window.addEventListener('message', (event) => {
        if (event.source !== window.parent) return;
        const { type, message } = event.data;
        window.parent.postMessage({ type, message }, '*');
      });

      window.console = new Proxy(console, {
        get(_, prop) {
          return (...args) => {
            if (typeof prop === 'string' && ['log', 'error', 'warn'].includes(prop)) {
              window.parent.postMessage({ type: prop, message: args.map(arg => arg?.toString?.() || arg) }, '*');
            }
          };
        }
      });

      window.addEventListener('error', (event) => {
        event.preventDefault();
        window.parent.postMessage({
          type: 'error',
          message: [\`Syntax Error: \${ event.message } at \${ event.filename }:\${ event.lineno }:\${ event.colno }\`]
        }, '*');
      });

      try {
        ${jsCode.value}
      } catch (e) {
        console.error(e);
      }
    <\/script>
  `;
};

const updateIframe = () => {
  if (iframeRef.value) {
    iframeRef.value.srcdoc = generateIframeContent();
  }
};

const handleConsoleMessages = (event: MessageEvent) => {
  if (event.data?.type) {
    logs.value.push({ type: event.data.type, message: event.data.message });
  }
};

watch(logs, async () => {
  await nextTick();
  consoleContainer.value?.scrollTo({ top: consoleContainer.value.scrollHeight, behavior: 'smooth' });
}, { deep: true, flush: 'post' });

onMounted(() => {
  window.addEventListener('message', handleConsoleMessages);
  updateIframe();
});

watch(htmlCode, updateIframe);
</script>

<template>
  <div class="w-full h-full flex flex-col">
    <button @click="updateIframe" class="bg-blue-500 text-white py-2">Run</button>
    <div class="flex-grow">
      <TheSplit direction="vertical" :initialPercentage="50">
        <template #first>
          <TheSplit direction="horizontal" :initialPercentage="50">
            <template #first>
              <div class="p-4 h-full flex flex-col">
                <span>HTML</span>
                <textarea v-model="htmlCode"
                  class="w-full h-full p-2 font-mono border border-gray-300 rounded resize-none"></textarea>
              </div>
            </template>
            <template #second>
              <div class="p-4 h-full flex flex-col">
                <span>JavaScript</span>
                <textarea v-model="jsCode"
                  class="w-full h-full p-2 font-mono border border-gray-300 rounded resize-none"></textarea>
              </div>
            </template>
          </TheSplit>
        </template>
        <template #second>
          <TheSplit direction="horizontal" :initialPercentage="55">
            <template #first>
              <div class="p-4 h-full flex flex-col">
                <span>Preview</span>
                <div class="w-full h-full border border-gray-300 rounded overflow-hidden">
                  <iframe ref="iframeRef" class="w-full h-full border-none" sandbox="allow-scripts"></iframe>
                </div>
              </div>
            </template>
            <template #second>
              <div class="p-4 h-full flex flex-col">
                <span>Console</span>
                <div ref="consoleContainer"
                  class="w-full h-full p-2 border border-gray-300 rounded overflow-y-auto bg-gray-100 font-mono">
                  <div v-for="(log, index) in logs" :key="index"
                    :class="{ 'text-red-500': log.type === 'error', 'text-yellow-500': log.type === 'warn' }">
                    [{{ log.type.toUpperCase() }}] {{ log.message.join(' ') }}
                  </div>
                </div>
              </div>
            </template>
          </TheSplit>
        </template>
      </TheSplit>
    </div>
  </div>
</template>
