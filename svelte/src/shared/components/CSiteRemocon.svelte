<svelte:options customElement={{ tag: 'c-site-remocon', shadow: 'none' }} />

<script lang="ts">
  import { mdiChevronDown, mdiChevronUp, mdiWeatherNight } from '@mdi/js'
  import { onMount } from 'svelte'

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
    if (stored === 'dark') {
      isDark = true
      applyDarkClasses(true)
      return
    }
    if (stored === 'light') {
      isDark = false
      applyDarkClasses(false)
      return
    }

    if (window.matchMedia?.('(prefers-color-scheme: dark)').matches) {
      isDark = true
      applyDarkClasses(true)
    }
  })
</script>

<div class="fixed bottom-0 right-0 z-40 print:hidden">
  <div class="z-50 flex gap-1 text-white opacity-80">
    <button type="button" class="cursor-pointer rounded bg-[#8888] p-1.5 dark:text-yellow-500" on:click={toggleDark}>
      <ZIcon path={mdiWeatherNight} size={24} />
    </button>
    <button type="button" class="cursor-pointer rounded bg-[#8888] p-1.5" on:click={scrollToTop}>
      <ZIcon path={mdiChevronUp} size={24} />
    </button>
    <button type="button" class="cursor-pointer rounded bg-[#8888] p-1.5" on:click={scrollToBottom}>
      <ZIcon path={mdiChevronDown} size={24} />
    </button>
  </div>
</div>
