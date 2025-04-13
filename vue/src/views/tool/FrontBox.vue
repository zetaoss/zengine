<script setup lang="ts">
import { ref, onMounted } from 'vue';

import TheConsole from '@common/components/console/TheConsole.vue';
import { type Log } from '@common/components/console/utils';

declare global {
  interface Window {
    console: Console;
  }
}

const htmlCode = ref('<h1>Hello, World!</h1>');
const jsCode = ref(`console.log('hello');`);
const iframe = ref<HTMLIFrameElement | null>(null);
const logs = ref<Log[]>([]);

function getContent(): string {
  const sanitizedHtml = htmlCode.value.trim();
  const wrappedHtml = sanitizedHtml.startsWith('<html>') ? sanitizedHtml : `<html><body>${sanitizedHtml}</body></html>`;
  const fullHtml = `${wrappedHtml}<script>try{new Function(\`${jsCode.value.replace(/\\/g, '\\\\')}\`)();}catch(e){console.error(\`Uncaught \${e.name}: \${e.message} \${e.stack.split('\\n')[2]}\`);}<\/script>`;
  return fullHtml;
}

const run = () => {
  logs.value = [];
  if (iframe.value) {
    const { contentDocument: doc, contentWindow: win } = iframe.value;
    if (doc && win) {
      win.console = new Proxy(console, {
        get(_, p) {
          if (p === 'constructor') {
            return { name: 'console' };
          }
          return (...args: unknown[]) => {
            const level = p as string
            logs.value.push({ level, args })
          }
        }
      })
      doc.open();
      doc.write(getContent());
      doc.close();
    }
  }
};

onMounted(() => {
  run();
});
</script>

<template>
  <div class="grid grid-cols-1 md:grid-cols-2 gap-4 p-4">
    <div>
      <div class="py-2">
        <b>Playground</b>
        <button @click="run" class="ml-2 bg-blue-500 dark:bg-blue-400 text-white dark:text-gray-900 text-sm font-semibold py-1.5 px-4 rounded
                   hover:bg-blue-600 dark:hover:bg-blue-500 transition">
          Run
        </button>
      </div>
      <textarea v-model="htmlCode" class="w-full h-[35vh]" />
      <textarea v-model="jsCode" class="w-full h-[35vh]" />
    </div>
    <div>
      <div class="bg-white h-[50vh]">
        <iframe ref="iframe" class="w-full h-full border-none" />
      </div>
      <div class="text-center font-bold bg-slate-400 dark:bg-slate-600 text-white">Console</div>
      <div class="h-[30vh] overflow-hidden overflow-y-scroll bg-slate-300 dark:bg-slate-800">
        <TheConsole :logs="logs" />
      </div>
    </div>
  </div>
</template>
