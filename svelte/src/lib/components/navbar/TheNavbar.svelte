<svelte:options runes={true} />

<script lang="ts">
  import { resolve } from '$app/paths'
  import { page } from '$app/state'
  import CSearch from '$shared/components/navbar/CSearch.svelte'

  import UserMenu from './UserMenu.svelte'

  let path = $derived(page.url.pathname)
  let isWiki = $derived(path === '/' || path.startsWith('/wiki') || path.startsWith('/onelines'))
  let isForum = $derived(path.startsWith('/forum'))
  let isTool = $derived(path.startsWith('/tool'))
</script>

<nav id="navbar" class="bg-slate-700 text-sm text-white">
  <div class="mx-auto flex w-full max-w-8xl flex-wrap items-center">
    <div class="order-1 flex h-12 items-stretch">
      <a class="navlink" href={resolve('/')}>
        <img alt="zetawiki" src="/zeta.svg?1" class="h-6 w-6" />
      </a>
      <a class={`navlink ${isWiki ? 'text-yellow-200!' : ''}`} href={resolve('/')}> 위키 </a>
      <a class={`navlink ${isForum ? 'text-yellow-200!' : ''}`} href={resolve('/forum')}> 포럼 </a>
      <a class={`navlink ${isTool ? 'text-yellow-200!' : ''}`} href={resolve('/tool/common-report')}> 도구 </a>
    </div>
    <div class="order-2 contents md:ml-auto md:flex md:w-auto md:max-w-2xl md:flex-1 md:items-stretch">
      <div class="order-2 ml-auto contents md:group md:relative md:inline-block">
        <UserMenu />
      </div>
      <div class="order-3 basis-full min-w-0 md:order-1 md:basis-auto md:flex-1">
        {#if isWiki}
          <CSearch />
        {/if}
      </div>
    </div>
  </div>
</nav>
