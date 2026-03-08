<svelte:options customElement={{ tag: 'c-adsense-slot', shadow: 'none' }} />

<script lang="ts">
  import { onMount, tick } from 'svelte'

  import { getTrackingState, pushAds } from '$shared/stores/trackingStore'
  import { getZConf } from '$shared/utils/zConf'

  export let index: 0 | 1

  let mounted = false
  let renderSeq = 0
  let adClient = ''
  let adSlot = ''
  let adInsEl: HTMLElement | null = null
  let pushed = false
  let show = false

  function block() {
    show = false
    adClient = ''
    adSlot = ''
    pushed = false
  }

  async function render() {
    if (!mounted) return
    const seq = ++renderSeq
    if (!getTrackingState().canShowAds) return block()

    const zConf = getZConf()
    const client = zConf.adClient || ''
    const slot = zConf.adSlots[index]
    if (!client || !slot) return block()

    const changed = adClient !== client || adSlot !== slot

    adClient = client
    adSlot = slot
    if (changed) pushed = false
    show = true
    await tick()

    if (seq !== renderSeq) return
    if (!adInsEl) return block()

    const status = adInsEl.getAttribute('data-adsbygoogle-status')
    if (pushed || status === 'done' || status === 'filled') return

    if (!pushAds()) {
      if (seq !== renderSeq) return
      return block()
    }

    if (seq !== renderSeq) return
    pushed = true
  }

  onMount(() => {
    mounted = true
    void render()
    const onAdsenseReady = () => {
      void render()
    }
    window.addEventListener('adsense:ready', onAdsenseReady)
    return () => {
      mounted = false
      window.removeEventListener('adsense:ready', onAdsenseReady)
      block()
    }
  })

  $: {
    void index
    if (mounted) void render()
  }
</script>

{#if show}
  <div class="mx-auto w-full max-w-180">
    <ins
      bind:this={adInsEl}
      class="adsbygoogle block w-full!"
      data-ad-client={adClient}
      data-ad-slot={adSlot}
      data-ad-format="horizontal"
      data-full-width-responsive="true"
    ></ins>
  </div>
{/if}

<style>
  :global(c-adsense-slot) {
    display: block;
    width: 100%;
  }
</style>
