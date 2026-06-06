<svelte:options runes={true} />

<script lang="ts">
  import { goto } from '$app/navigation'
  import { resolve } from '$app/paths'
  import { page } from '$app/state'
  import ZTabs from '$shared/ui/ZTabs.svelte'

  let { children } = $props()

  const tabs = [
    { value: 'tasks', label: '작업목록', href: resolve('/tool/ai-edit/tasks') },
    { value: 'prompts', label: '프롬프트', href: resolve('/tool/ai-edit/prompts') },
  ]

  let tab = $derived.by(() => {
    if (page.url.pathname.startsWith('/tool/ai-edit/prompts')) return 'prompts'
    return 'tasks'
  })

  function setTab(nextTab: string) {
    const p: '/tool/ai-edit/prompts' | '/tool/ai-edit/tasks' = nextTab === 'prompts' ? '/tool/ai-edit/prompts' : '/tool/ai-edit/tasks'
    void goto(resolve(p), { replaceState: true, noScroll: true })
  }
</script>

<div class="p-5">
  <h2 class="my-5 text-2xl font-bold">AI 편집</h2>
  <ZTabs {tabs} selected={tab} onChange={setTab} />
</div>

{@render children()}
