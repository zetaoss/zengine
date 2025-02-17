<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue';
import TheSplit from '@/components/TheSplit.vue';

const htmlCode = ref('<h1>Hello, World!</h1>');
const jsCode = ref('console.log(new Date());');
const iframeRef = ref<HTMLIFrameElement | null>(null);
const logs = ref<{ type: string; message: unknown[] }[]>([]);
const consoleContainer = ref<HTMLElement | null>(null);
const autoUpdate = ref(true);

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

onMounted(() => {
  window.addEventListener('message', handleConsoleMessages);
  updateIframe();
});

watch(logs, async () => {
  await nextTick();
  consoleContainer.value?.scrollTo({ top: consoleContainer.value.scrollHeight, behavior: 'smooth' });
}, { deep: true, flush: 'post' });

watch(autoUpdate, (newValue) => {
  if (newValue) {
    updateIframe();
  }
});

watch(htmlCode, () => {
  if (autoUpdate.value) {
    updateIframe();
  }
});
</script>

<template>
  <div class="w-full h-full flex flex-col">

    <div
      class="flex items-center py-2 px-4 bg-gray-100 dark:bg-gray-800 border-b border-gray-300 dark:border-gray-700 transition-colors">
      <div class="ml-auto">
        <!-- Run Button -->
        <button @click="updateIframe" class="bg-blue-500 dark:bg-blue-400 text-white dark:text-gray-900 text-sm font-semibold py-1.5 px-4 rounded
                   hover:bg-blue-600 dark:hover:bg-blue-500 transition">
          Run
        </button>
      </div>
    </div>

    <div class="flex-grow">
      <TheSplit direction="vertical" :initialPercentage="50">
        <template #first>
          <TheSplit direction="horizontal" :initialPercentage="50">
            <template #first>
              <div class="p-4 h-full flex flex-col">
                <div class="flex items-center mb-2">
                  <span class="text-gray-700 dark:text-gray-300">HTML</span>
                  <div class="ml-auto flex items-center space-x-4">
                    <!-- Auto Update Toggle -->
                    <label class="flex items-center cursor-pointer">
                      <input type="checkbox" v-model="autoUpdate" class="hidden">
                      <span class="relative w-10 h-5 rounded-full transition-all duration-300"
                        :class="autoUpdate ? 'bg-blue-600 dark:bg-blue-400' : 'bg-gray-300 dark:bg-gray-600'">
                        <span
                          class="absolute left-1 top-1 w-3 h-3 bg-white dark:bg-gray-200 rounded-full transition-all duration-300"
                          :class="autoUpdate ? 'translate-x-5' : ''"></span>
                      </span>
                      <span class="ml-2 text-sm text-gray-700 dark:text-gray-300">Auto Update</span>
                    </label>
                  </div>
                </div>
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
              <!-- Console Section -->
              <div class="p-4 h-full flex flex-col relative">
                <div class="flex justify-between items-center mb-2">
                  <span class="text-gray-700 dark:text-gray-300">Console</span>
                  <button @click="logs = []" class="text-xs px-2 py-1 rounded transition border border-gray-300 dark:border-gray-600
                   text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700">
                    Clear
                  </button>
                </div>
                <div ref="consoleContainer"
                  class="w-full h-full p-2 border border-gray-300 dark:border-gray-600 rounded overflow-y-auto bg-gray-100 dark:bg-gray-800 font-mono">
                  <div v-for="(log, index) in logs" :key="index" :class="{
                    'text-red-500 dark:text-red-400': log.type === 'error',
                    'text-yellow-500 dark:text-yellow-400': log.type === 'warn'
                  }">
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
