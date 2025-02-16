<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { QSplitter } from 'quasar';
import Card from '@common/components/ui/UICard.vue';
import CardContent from '@common/components/ui/UICardContent.vue';
import Button from '@common/components/ui/UIButton.vue';
const htmlCode = ref('<h1>Hello, World!</h1>');
const jsCode = ref('console.log("Hello from JS");');
const iframeRef = ref<HTMLIFrameElement | null>(null);
const logs = ref<{ type: string; message: unknown[] }[]>([]);

const splitterModel1 = ref(50);
const splitterModel2 = ref(50);
const splitterModel3 = ref(70);

const updateIframe = () => {
  if (iframeRef.value) {
    const iframe = iframeRef.value;
    const doc = iframe.contentDocument || iframe.contentWindow?.document;
    if (doc) {
      doc.open();
      doc.write(`
        <html>
          <head>
            <script>
              window.addEventListener('message', (event) => {
                if (event.source !== window.parent) return;
                const { type, message } = event.data;
                window.parent.postMessage({ type, message }, '*');
              });

              window.console = new Proxy({}, {
                get(_, prop) {
                  return (...args) => {
                    if (typeof prop === 'string' && ['log', 'error', 'warn'].includes(prop)) {
                      parent.postMessage({ type: prop, message: args.map(arg => arg && arg.toString ? arg.toString() : arg) }, '*');
                    }
                  };
                }
              });

              window.addEventListener('error', (event) => {
                event.preventDefault();
                parent.postMessage({ type: 'error', message: [\`Syntax Error: \${ event.message } at \${ event.filename }: \${ event.lineno }: \${ event.colno }\`] }, '*');
              });
            <\/script>
</head>

<body>
            ${htmlCode.value}
            <script>
              setTimeout(() => {
                try {
                  ${jsCode.value}
                } catch (e) {
                  console.error(e);
                }
              }, 10);
            <\/script>
          </body>

</html>
`);
      doc.close();
    }
  }
};

const handleConsoleMessages = (event: MessageEvent) => {
  if (event.data && event.data.type) {
    logs.value.push({ type: event.data.type, message: event.data.message });
  }
};

onMounted(() => {
  window.addEventListener('message', handleConsoleMessages);
  updateIframe();
});

watch([htmlCode, jsCode], updateIframe);
</script>

<template>
  <QSplitter v-model="splitterModel1" class="h-screen">
    <template v-slot:before>
      <QSplitter v-model="splitterModel2" vertical>
        <template v-slot:before>
          <div class="p-2">
            <h2 class='text-xl mb-2'>HTML</h2>
            <textarea v-model='htmlCode' class='w-full h-32 p-2 border rounded'></textarea>
          </div>
        </template>
        <template v-slot:after>
          <div class="p-2">
            <h2 class='text-xl mb-2'>JavaScript</h2>
            <textarea v-model='jsCode' class='w-full h-32 p-2 border rounded'></textarea>
            <Button class='mt-4' @click='updateIframe'>Run</Button>
          </div>
        </template>
      </QSplitter>
    </template>
    <template v-slot:after>
      <QSplitter v-model="splitterModel3" vertical>
        <template v-slot:before>
          <div class="p-2">
            <h2 class='text-xl mb-2'>Preview</h2>
            <Card>
              <CardContent>
                <iframe ref='iframeRef' class='w-full h-64 border rounded'
                  sandbox="allow-scripts allow-same-origin"></iframe>
              </CardContent>
            </Card>
          </div>
        </template>
        <template v-slot:after>
          <div class="p-2">
            <h2 class='text-xl mb-2'>Console Output</h2>
            <Card>
              <CardContent class='h-32 overflow-auto p-2 bg-black text-white'>
                <div v-for='(log, index) in logs' :key='index'
                  :class="{ 'text-red-500': log.type === 'error', 'text-yellow-500': log.type === 'warn' }">
                  [{{ log.type.toUpperCase() }}] {{ log.message.join(' ') }}
                </div>
              </CardContent>
            </Card>
          </div>
        </template>
      </QSplitter>
    </template>
  </QSplitter>
</template>
