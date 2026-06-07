<svelte:options customElement={{ tag: 'c-site-remocon', shadow: 'none' }} />

<script lang="ts">
  import { mdiChevronDown, mdiChevronUp, mdiWeatherNight } from '@mdi/js'
  import { onMount } from 'svelte'

  import CButton from '$shared/ui/CButton.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import { scrollToBottom, scrollToTop } from '$shared/utils/scroll'

  let isDark = false

  const applyDarkClasses = (enabled: boolean) => {
    document.documentElement.classList.toggle('dark', enabled)
    document.documentElement.classList.toggle('skin-theme-clientpref-night', enabled)
  }

  const toggleDark = () => {
    isDark = !isDark
    applyDarkClasses(isDark)
    localStorage.setItem('theme', isDark ? 'dark' : 'light')
  }

  onMount(() => {
    const stored = localStorage.getItem('theme')
    const prefersDark = window.matchMedia?.('(prefers-color-scheme: dark)').matches ?? false
    isDark = stored ? stored === 'dark' : prefersDark
    applyDarkClasses(isDark)
  })
</script>

<div class="fixed bottom-0 right-0 z-40 print:hidden p-2 text-always-white">
  <div class="z-50 flex gap-1 opacity-60">
    <CButton size="icon-lg" class="bg-[#8888]! border-0! text-always-white! dark:text-[#eb0]!" onclick={toggleDark}>
      <ZIcon path={mdiWeatherNight} class="size-6" />
    </CButton>
    <CButton size="icon-lg" class="bg-[#8888]! border-0! text-always-white!" onclick={scrollToTop}>
      <ZIcon path={mdiChevronUp} class="size-6" />
    </CButton>
    <CButton size="icon-lg" class="bg-[#8888]! border-0! text-always-white!" onclick={scrollToBottom}>
      <ZIcon path={mdiChevronDown} class="size-6" />
    </CButton>
  </div>
</div>
