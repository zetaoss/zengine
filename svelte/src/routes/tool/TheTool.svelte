<svelte:options runes={true} />

<script lang="ts">
  import type { Snippet } from 'svelte'

  import { resolve } from '$app/paths'
  import { page } from '$app/state'

  interface NavItem {
    href: string
    key: string
    label: string
    external?: boolean
  }

  const navItems: NavItem[] = [
    { key: '/tool/binder', href: resolve('/tool/binder'), label: '바인더' },
    { key: '/tool/common-report', href: resolve('/tool/common-report'), label: '통용' },
    { key: '/tool/article-tpl', href: resolve('/tool/article-tpl'), label: '템플릿' },
    { key: '/tool/write-request', href: resolve('/tool/write-request'), label: '작성요청' },
    { key: '/tool/ai-edit', href: resolve('/tool/ai-edit'), label: 'AI 편집' },
    { key: '/tool/stat', href: resolve('/tool/stat'), label: '통계' },
    { key: '/tool/frontplay', href: resolve('/tool/frontplay'), label: 'FrontPlay' },
  ]

  function isActive(pathname: string, to: string) {
    return pathname === to || pathname.startsWith(`${to}/`)
  }

  let pendingKey = $state('')
  let pathname = $derived(page.url.pathname)
  let { children }: { children?: Snippet } = $props()

  $effect(() => {
    if (pendingKey && isActive(pathname, pendingKey)) {
      pendingKey = ''
    }
  })
</script>

<div class="h-full grid md:grid-cols-[140px_1fr]">
  <aside class="z-bg-muted h-full p-4">
    <table class="w-full">
      <thead>
        <tr>
          <th class="text-left text-lg font-bold">도구</th>
        </tr>
      </thead>
      <tbody>
        {#each navItems as item (item.key)}
          <tr>
            <td>
              <a
                href={item.href}
                class={`block w-full rounded px-3 py-2 text-left transition text-foreground hover:no-underline hover:bg-a-gray-200 ${
                  isActive(pathname, item.key) || pendingKey === item.key ? 'bg-a-gray-200 font-semibold' : ''
                }`}
                rel={item.external ? 'external' : undefined}
                data-sveltekit-reload={item.external ? true : undefined}
                onpointerdown={(event) => {
                  if (item.external) return
                  if (event.button !== 0) return
                  if (event.ctrlKey || event.metaKey || event.shiftKey || event.altKey) return
                  pendingKey = item.key
                }}
              >
                {item.label}
              </a>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </aside>

  <div>
    {@render children?.()}
  </div>
</div>
