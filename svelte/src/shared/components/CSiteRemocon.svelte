<svelte:options customElement={{ tag: 'c-site-remocon', shadow: 'none' }} />

<script lang="ts">
  import { mdiChevronDown, mdiChevronUp, mdiWeatherNight } from '@mdi/js'
  import { onMount } from 'svelte'

  import ZIcon from '$shared/ui/ZIcon.svelte'
  import { scrollToBottom, scrollToTop } from '$shared/utils/scroll'

  let isDark = false

  const toggleDark = () => {
    isDark = !isDark
    document.documentElement.classList.toggle('dark', isDark)
    localStorage.setItem('theme', isDark ? 'dark' : 'light')
  }

  onMount(() => {
    const stored = localStorage.getItem('theme')
    if (stored === 'dark') {
      isDark = true
      document.documentElement.classList.add('dark')
      return
    }
    if (stored === 'light') {
      isDark = false
      document.documentElement.classList.remove('dark')
      return
    }

    if (window.matchMedia?.('(prefers-color-scheme: dark)').matches) {
      isDark = true
      document.documentElement.classList.add('dark')
    }
  })
</script>

<div class="fixed bottom-0 right-0 z-40 print:hidden">
  <div class="z-50 mx-0.5 text-white opacity-80">
    <button type="button" class="ml-0.5 rounded bg-[#8888] p-1.5" class:text-yellow-500={isDark} on:click={toggleDark}>
      <ZIcon path={mdiWeatherNight} size={24} />
    </button>
    <button type="button" class="ml-0.5 rounded bg-[#8888] p-1.5" on:click={scrollToTop}>
      <ZIcon path={mdiChevronUp} size={24} />
    </button>
    <button type="button" class="ml-0.5 rounded bg-[#8888] p-1.5" on:click={scrollToBottom}>
      <ZIcon path={mdiChevronDown} size={24} />
    </button>
  </div>
</div>
