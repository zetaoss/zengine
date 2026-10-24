<svelte:options runes={true} />

<script lang="ts">
  import type { Snippet } from 'svelte'

  import { resolve } from '$app/paths'
  import { page } from '$app/state'
  import useAuthStore from '$lib/stores/auth'
  import ZTabs from '$lib/ui/ZTabs.svelte'

  let { children }: { children?: Snippet } = $props()

  const { userInfo } = useAuthStore()
  let isSysop = $derived(($userInfo?.groups ?? []).includes('sysop'))

  const tabs = [
    { value: 'define', label: '문서정의', href: resolve('/tool/ai/define') },
    { value: 'eval', label: '문서평가', href: resolve('/tool/ai/eval') },
    { value: 'chains', label: '체인', href: resolve('/tool/ai/chains') },
    { value: 'models', label: '모델', href: resolve('/tool/ai/models') },
  ]
  // A prompt template's pages belong to the chains tab, which lists them under their chain.
  let selected = $derived(
    page.url.pathname.startsWith(resolve('/tool/ai/prompts'))
      ? 'chains'
      : (tabs.find((t) => page.url.pathname.startsWith(t.href))?.value ?? 'eval'),
  )
</script>

{#if isSysop}
  <div class="px-4 pt-4">
    <ZTabs {tabs} {selected} />
  </div>
  {@render children?.()}
{:else}
  <div class="p-8 text-center text-muted-foreground">관리자 전용입니다.</div>
{/if}
