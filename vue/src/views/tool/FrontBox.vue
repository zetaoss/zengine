<!-- FrontBox.vue -->
<script setup lang="ts">
import { ref } from 'vue'

import ZButton from '@common/ui/ZButton.vue'
import { SandboxFrame, SandboxConsole, type SandboxLog } from '@common/components/sandbox'

const htmlCode = ref(`<h1>Hello, World!</h1>

<script>
console.log("Hello HTML");
<\/script>`)

const jsCode = ref(`console.log('Hello JS');`)

const logs = ref<SandboxLog[]>([])
const sandboxRef = ref<InstanceType<typeof SandboxFrame> | null>(null)

const run = () => {
  sandboxRef.value?.run()
}

const updateLogs = (newLogs: SandboxLog[]) => {
  logs.value = newLogs
}
</script>

<template>
  <div class="grid grid-cols-1 md:grid-cols-2 gap-4 p-4">
    <div>
      <div class="py-2 flex items-center">
        <b>Playground</b>
        <ZButton class="ml-2" @click="run">Run</ZButton>
      </div>

      <div class="space-y-2">
        <textarea v-model="htmlCode" class="h-[35vh] w-full resize-none font-mono border rounded p-2" />
        <textarea v-model="jsCode" class="h-[35vh] w-full resize-none font-mono border rounded p-2" />
      </div>
    </div>

    <div class="flex flex-col gap-2">
      <div class="h-[50vh] border rounded overflow-hidden">
        <SandboxFrame ref="sandboxRef" :html="htmlCode" :js="jsCode" class="w-full h-full" @update:logs="updateLogs" />
      </div>

      <div class="flex-1 flex flex-col min-h-0">
        <header class="text-center font-bold bg-slate-400 dark:bg-slate-600 text-white py-1">
          Console
        </header>
        <div class="h-[30vh] overflow-y-auto bg-[var(--console-bg)]">
          <SandboxConsole :logs="logs" />
        </div>
      </div>
    </div>
  </div>
</template>
