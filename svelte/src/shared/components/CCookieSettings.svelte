<script lang="ts">
  import { mdiChevronDown } from '@mdi/js'

  import { getTrackingState } from '$shared/stores/trackingStore'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZToggle from '$shared/ui/ZToggle.svelte'
  import { deleteCookie, writeCookie } from '$shared/utils/cookie'

  export let open = false
  export let onClose: () => void = () => {}

  let analyticsStorage = false
  let marketingStorage = false
  let necessaryExpanded = false
  let analyticsExpanded = false
  let marketingExpanded = false

  type ConsentState = {
    analytics_storage: 'granted' | 'denied'
    ad_storage: 'granted' | 'denied'
    ad_user_data: 'granted' | 'denied'
    ad_personalization: 'granted' | 'denied'
  }

  function writeConsent(state: ConsentState) {
    const value = new URLSearchParams({
      analytics_storage: state.analytics_storage,
      ad_storage: state.ad_storage,
      ad_user_data: state.ad_user_data,
      ad_personalization: state.ad_personalization,
    }).toString()
    writeCookie('consent', value)
    const legacyKeys = ['analytics_storage', 'ad_storage', 'ad_user_data', 'ad_personalization']
    legacyKeys.forEach((key) => deleteCookie(key))
  }

  function updateGoogleConsent(state: ConsentState) {
    if (typeof window.gtag !== 'function') return
    window.gtag('consent', 'update', {
      ad_storage: state.ad_storage,
      ad_user_data: state.ad_user_data,
      ad_personalization: state.ad_personalization,
      analytics_storage: state.analytics_storage,
    })
  }

  function syncFromCookies() {
    const state = getTrackingState()
    analyticsStorage = state.consentAnalytics
    marketingStorage = state.consentAds
  }

  function closeConsent() {
    onClose()
  }

  function allowSelection() {
    const state: ConsentState = {
      analytics_storage: analyticsStorage ? 'granted' : 'denied',
      ad_storage: marketingStorage ? 'granted' : 'denied',
      ad_user_data: marketingStorage ? 'granted' : 'denied',
      ad_personalization: marketingStorage ? 'granted' : 'denied',
    }
    writeConsent(state)
    updateGoogleConsent(state)
    window.dispatchEvent(new Event('consent:updated'))
    window.location.reload()
  }

  function allowNecessaryOnly() {
    analyticsStorage = false
    marketingStorage = false
    allowSelection()
  }

  function allowAllConsent() {
    analyticsStorage = true
    marketingStorage = true
    allowSelection()
  }

  function toggleNecessary() {
    necessaryExpanded = !necessaryExpanded
  }

  function toggleAnalytics() {
    analyticsExpanded = !analyticsExpanded
  }

  function toggleMarketing() {
    marketingExpanded = !marketingExpanded
  }

  $: if (open) {
    syncFromCookies()
  }
</script>

{#if open}
  <div
    class="fixed inset-0 z-60 flex items-end justify-end bg-black/30 p-4"
    role="button"
    tabindex="0"
    on:click|self={closeConsent}
    on:keydown={(e) => (e.key === 'Escape' || e.key === 'Enter' || e.key === ' ') && closeConsent()}
  >
    <div
      class="relative w-full max-w-xl rounded border border-neutral-200/70 bg-white p-5 text-left text-neutral-900 shadow-xl dark:border-zinc-700/60 dark:bg-zinc-900 dark:text-neutral-100"
    >
      <button
        aria-label="Close consent dialog"
        class="absolute right-4 top-4 cursor-pointer text-neutral-500 hover:text-neutral-700 dark:text-neutral-400 dark:hover:text-neutral-200"
        type="button"
        on:click={closeConsent}>✕</button
      >
      <div class="flex max-h-[calc(100vh-5rem)] flex-col">
        <div class="mb-3 flex items-center gap-2 border-b border-neutral-200 pb-2 dark:border-slate-600">
          <img alt="logo" class="h-7 w-7" src="/zeta.svg" />
          <span class="text-sm font-bold tracking-wide text-neutral-700 dark:text-neutral-200">zetawiki.com</span>
        </div>
        <div class="z-scrollbar flex-1 overflow-y-auto pr-1">
          <h2 class="mb-2 text-lg font-semibold">Cookie Consent</h2>
          <p class="mb-4 text-sm text-neutral-600 dark:text-neutral-300">
            When you visit any of our websites, it may store or retrieve information on your browser, mostly in the form of cookies. This
            information might be about you, your preferences, or your device and is mostly used to make the site work as you expect it to.
            The information does not usually directly identify you, but it can give you a more personalized experience. Because we respect
            your right to privacy, you can choose not to allow some types of cookies. Click on the different category headings to find out
            more and manage your preferences. Please note, blocking some types of cookies may impact your experience of the site and the
            services we are able to offer.
          </p>
          <h3 class="mb-3 text-sm font-bold tracking-wide text-neutral-700 dark:text-neutral-200">Manage Consent Preferences</h3>

          <div class="space-y-3">
            <div class="rounded border border-neutral-200 dark:border-slate-600">
              <div class="flex items-center gap-3 p-3">
                <button
                  class="flex min-w-0 flex-1 cursor-pointer items-center gap-2 text-left font-semibold text-neutral-800 dark:text-neutral-100"
                  type="button"
                  aria-expanded={necessaryExpanded}
                  aria-controls="necessary-desc"
                  on:click={toggleNecessary}
                >
                  <span
                    class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-slate-200 text-neutral-600 dark:bg-slate-600 dark:text-neutral-200"
                  >
                    <ZIcon path={mdiChevronDown} size={14} class={`transition-transform ${necessaryExpanded ? 'rotate-180' : ''}`} />
                  </span>
                  <span>Strictly necessary cookies</span>
                </button>
                <span class="text-sm font-bold text-neutral-700 dark:text-neutral-200">Always active</span>
              </div>
              {#if necessaryExpanded}
                <div
                  id="necessary-desc"
                  class="border-t border-neutral-200 px-3 pb-3 pt-2 text-sm text-neutral-600 dark:border-slate-600 dark:text-neutral-300"
                >
                  Necessary cookies are required for basic features like secure login, session handling, and consent preference storage.
                </div>
              {/if}
            </div>

            <div class="rounded border border-neutral-200 dark:border-slate-600">
              <div class="flex items-center gap-3 p-3">
                <button
                  class="flex min-w-0 flex-1 cursor-pointer items-center gap-2 text-left font-semibold text-neutral-800 dark:text-neutral-100"
                  type="button"
                  aria-expanded={analyticsExpanded}
                  aria-controls="analytics-desc"
                  on:click={toggleAnalytics}
                >
                  <span
                    class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-slate-200 text-neutral-600 dark:bg-slate-600 dark:text-neutral-200"
                  >
                    <ZIcon path={mdiChevronDown} size={14} class={`transition-transform ${analyticsExpanded ? 'rotate-180' : ''}`} />
                  </span>
                  <span>Performance and analytics cookies</span>
                </button>
                <ZToggle bind:checked={analyticsStorage} label="analytics_storage" />
              </div>
              {#if analyticsExpanded}
                <div
                  id="analytics-desc"
                  class="border-t border-neutral-200 px-3 pb-3 pt-2 text-sm text-neutral-600 dark:border-slate-600 dark:text-neutral-300"
                >
                  Analytics cookies help us measure traffic and understand how the site is used so we can improve performance.
                </div>
              {/if}
            </div>

            <div class="rounded border border-neutral-200 dark:border-slate-600">
              <div class="flex items-center gap-3 p-3">
                <button
                  class="flex min-w-0 flex-1 cursor-pointer items-center gap-2 text-left font-semibold text-neutral-800 dark:text-neutral-100"
                  type="button"
                  aria-expanded={marketingExpanded}
                  aria-controls="marketing-desc"
                  on:click={toggleMarketing}
                >
                  <span
                    class="inline-flex h-5 w-5 items-center justify-center rounded-full bg-slate-200 text-neutral-600 dark:bg-slate-600 dark:text-neutral-200"
                  >
                    <ZIcon path={mdiChevronDown} size={14} class={`transition-transform ${marketingExpanded ? 'rotate-180' : ''}`} />
                  </span>
                  <span>Marketing cookies</span>
                </button>
                <ZToggle bind:checked={marketingStorage} label="marketing_storage" />
              </div>
              {#if marketingExpanded}
                <div
                  id="marketing-desc"
                  class="border-t border-neutral-200 px-3 pb-3 pt-2 text-sm text-neutral-600 dark:border-slate-600 dark:text-neutral-300"
                >
                  Marketing cookies are used for ad personalization and campaign measurement across services.
                </div>
              {/if}
            </div>
          </div>
        </div>

        <div class="mt-4 grid grid-cols-3 gap-2 border-t border-neutral-200 pt-3 dark:border-slate-600">
          <button
            class="w-full cursor-pointer rounded bg-neutral-700 px-3 py-2.5 text-center text-sm font-bold text-white hover:bg-neutral-800 dark:bg-slate-300 dark:text-zinc-900 dark:hover:bg-slate-200"
            type="button"
            on:click={allowAllConsent}>Accept all cookies</button
          >
          <button
            class="w-full cursor-pointer rounded bg-neutral-700 px-3 py-2.5 text-center text-sm font-bold text-white hover:bg-neutral-800 dark:bg-slate-300 dark:text-zinc-900 dark:hover:bg-slate-200"
            type="button"
            on:click={allowNecessaryOnly}>Necessary cookies only</button
          >
          <button
            class="w-full cursor-pointer rounded bg-neutral-700 px-3 py-2.5 text-center text-sm font-bold text-white hover:bg-neutral-800 dark:bg-slate-300 dark:text-zinc-900 dark:hover:bg-slate-200"
            type="button"
            on:click={allowSelection}>Confirm my choices</button
          >
        </div>
      </div>
    </div>
  </div>
{/if}
