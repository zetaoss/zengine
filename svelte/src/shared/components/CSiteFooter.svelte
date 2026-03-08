<svelte:options customElement={{ tag: 'c-site-footer', shadow: 'none' }} />

<script lang="ts">
  import { onMount } from 'svelte'

  import CCookieSettings from '$shared/components/CCookieSettings.svelte'
  import { getTrackingState } from '$shared/stores/trackingStore'

  let consentOpen = false
  let showCookieSettings = false

  onMount(() => {
    const state = getTrackingState()
    const isStrict = state.policy === 'strict'
    showCookieSettings = isStrict
    consentOpen = isStrict && !state.hasConsentCookie
  })
</script>

<footer class="border-t bg-neutral-400 p-6 pb-9 text-center dark:bg-slate-900">
  <img alt="logo" class="mb-3 w-24" src="//storage.googleapis.com/zpub/logo.png" />
  <p class="mb-3 text-white">CC-BY-SA 3.0 · Powered by MediaWiki</p>

  <div class="mt-3">
    <a class="text-white" href="/wiki/제타위키:개인정보처리방침" rel="external" data-sveltekit-reload>개인정보처리방침</a>
    {#if showCookieSettings}
      ·
      <button class="cursor-pointer text-white" type="button" on:click={() => (consentOpen = true)}>쿠키설정</button>
    {/if}
    ·
    <a class="text-white" href="/wiki/제타위키" rel="external" data-sveltekit-reload>ABOUT</a>
  </div>

  {#if showCookieSettings}
    <CCookieSettings open={consentOpen} onClose={() => (consentOpen = false)} />
  {/if}
</footer>
