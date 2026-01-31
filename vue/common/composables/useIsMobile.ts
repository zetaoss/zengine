// @common/composables/useIsMobile.ts
import { onMounted, onUnmounted, ref } from 'vue'

export function useIsMobile() {
  const isMobile = ref(false)

  const update = () => {
    isMobile.value = window.innerWidth < 900
  }

  onMounted(() => {
    update()
    window.addEventListener('resize', update)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', update)
  })

  return { isMobile }
}
