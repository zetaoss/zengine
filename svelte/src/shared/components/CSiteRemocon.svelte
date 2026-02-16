<svelte:options customElement={{ tag: 'c-site-remocon', shadow: 'none' }} />

<script lang="ts">
  import { mdiChevronDown, mdiChevronUp, mdiWeatherNight } from '@mdi/js'
  import { onMount } from 'svelte'

  import ZButton from '$shared/ui/ZButton.svelte'
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

<div class="fixed bottom-0 right-0 z-40 print:hidden">
  <div class="z-50 flex gap-1 opacity-80">
    <ZButton class="bg-[#8888]! p-1.5! text-white! dark:text-yellow-500!" onclick={toggleDark}>
      <ZIcon path={mdiWeatherNight} size={24} />
    </ZButton>
    <ZButton class="bg-[#8888]! p-1.5! text-white!" onclick={scrollToTop}>
      <ZIcon path={mdiChevronUp} size={24} />
    </ZButton>
    <ZButton class="bg-[#8888]! p-1.5! text-white!" onclick={scrollToBottom}>
      <ZIcon path={mdiChevronDown} size={24} />
    </ZButton>
  </div>
</div>
