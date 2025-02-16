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
        get(target, prop) {
          return (...args) => {
            if (typeof prop === 'string' && ['log', 'error', 'warn'].includes(prop)) {
              window.parent.postMessage({ type: prop, message: args.map(arg => arg?.toString?.() || arg) }, '*');
            }
            return target[prop](...args);
          };
        }
      });

      window.addEventListener('error', (event) => {
        event.preventDefault();
        window.parent.postMessage({
          type: 'error',
          message: [\`Syntax Error: \${event.message} at \${event.filename}:\${event.lineno}:\${event.colno}\`]
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
  if (event.data && event.data.type) {
    logs.value.push({ type: event.data.type, message: event.data.message });
  }
};

watch(logs, async () => {
  await nextTick();
  if (consoleContainer.value) {
    consoleContainer.value.scrollTo({
      top: consoleContainer.value.scrollHeight,
      behavior: 'smooth',
    });
  }
}, { deep: true, flush: 'post' });

onMounted(() => {
  window.addEventListener('message', handleConsoleMessages);
  updateIframe();
});

watch([htmlCode, jsCode], updateIframe);
</script>

<template>
  <div class="w-full h-full flex flex-col">
    <div>aaaa</div>
    <div class="flex-grow">
      <TheSplit direction="vertical" :initialPercentage="50">
        <template #first>
          <TheSplit direction="horizontal" :initialPercentage="50">
            <template #first>
              <div class="section">
                <span>HTML</span>
                <textarea v-model="htmlCode" class="editor"></textarea>
              </div>
            </template>

            <template #second>
              <div class="section">
                <span>JavaScript</span>
                <textarea v-model="jsCode" class="editor"></textarea>
              </div>
            </template>
          </TheSplit>
        </template>

        <template #second>
          <TheSplit direction="horizontal" :initialPercentage="55"> <!-- 위/아래 분할 -->
            <template #first>
              <div class="section">
                <span>Preview</span>
                <div class="preview-container">
                  <iframe ref="iframeRef" class="iframe" sandbox="allow-scripts"></iframe>
                </div>
              </div>
            </template>

            <template #second>
              <div class="section">
                <span>Console</span>
                <div class="console-container" ref="consoleContainer">
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

<style scoped>
.frontbox {
  width: 100%;
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.section {
  padding: 1rem;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.editor {
  width: 100%;
  height: 100%;
  padding: 0.5rem;
  font-family: monospace;
  border: 1px solid #ccc;
  border-radius: 4px;
  resize: none;
}

.preview-container {
  width: 100%;
  height: 100%;
  border: 1px solid #ccc;
  border-radius: 4px;
  overflow: hidden;
}

.iframe {
  width: 100%;
  height: 100%;
  border: none;
}

.console-container {
  width: 100%;
  height: 100%;
  padding: 0.5rem;
  border: 1px solid #ccc;
  border-radius: 4px;
  overflow-y: auto;
  background-color: #f5f5f5;
  font-family: monospace;
}
</style>
