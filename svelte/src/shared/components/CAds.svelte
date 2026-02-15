<svelte:options customElement={{ tag: 'c-ads', shadow: 'none' }} />

<script lang="ts">
  import { onMount, tick } from 'svelte'

  export let client: string
  export let slot: string
  export let style: string | undefined

  const pushAd = () => {
    try {
      window.adsbygoogle = window.adsbygoogle || []
      window.adsbygoogle.push({})
    } catch {
      // ignore
    }
  }

  const render = async () => {
    await tick()
    pushAd()
  }

  onMount(() => {
    void render()
  })

  $: if (slot) {
    void render()
  }

  const defaultStyle = 'display:block;width:100%;height:100%;'
  $: insStyle = style ? `${defaultStyle}${style}` : defaultStyle
</script>

<ins class="adsbygoogle" style={insStyle} data-ad-client={client} data-ad-slot={slot}></ins>
