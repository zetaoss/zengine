<!-- ZModal.vue -->
<script setup lang="ts">
import ZButton from '@common/ui/ZButton.vue'
import ZIcon from '@common/ui/ZIcon.vue'
import { mdiClose } from '@mdi/js'
import { onMounted, onUnmounted } from 'vue'

const props = withDefaults(
  defineProps<{
    show: boolean
    title?: string
    okText?: string
    okColor?: 'ghost' | 'default' | 'danger' | 'primary'
    okDisabled?: boolean
    cancelText?: string
    closable?: boolean
    backdropClosable?: boolean
  }>(),
  {
    title: '',
    okText: '확인',
    okColor: 'danger',
    okDisabled: false,
    cancelText: '취소',
    closable: true,
    backdropClosable: true,
  },
)

const emit = defineEmits(['ok', 'cancel'])

function onKeyup(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.show && props.closable)
    emit('cancel')
}

onMounted(() => document.addEventListener('keyup', onKeyup))
onUnmounted(() => document.removeEventListener('keyup', onKeyup))
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 z-40 flex items-center justify-center bg-black/40"
      @click.self="backdropClosable && emit('cancel')">
      <div class="w-full max-w-[60vw] md:max-w-[40vw] border rounded-md bg-white dark:bg-gray-900">
        <ZButton v-if="closable" color="ghost" class="float-right m-1" @click="emit('cancel')">
          <ZIcon :path="mdiClose" />
        </ZButton>

        <header v-if="title" class="px-5 py-3 border-b">
          <h2 class="text-base font-semibold m-0">{{ title }}</h2>
        </header>

        <section class="p-5">
          <slot />
        </section>

        <footer class="flex justify-center gap-3 px-4 py-3 border-t">
          <ZButton :disabled="okDisabled" :color="okColor" @click="!okDisabled && emit('ok')">
            {{ okText }}
          </ZButton>
          <ZButton @click="emit('cancel')">
            {{ cancelText }}
          </ZButton>
        </footer>
      </div>
    </div>
  </Teleport>
</template>
