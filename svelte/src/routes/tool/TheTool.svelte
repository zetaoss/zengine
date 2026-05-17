<svelte:options runes={true} />

<script lang="ts">
  import type { Snippet } from 'svelte'

  import { resolve } from '$app/paths'
  import { page } from '$app/state'

  interface NavItem {
    to: '/tool/binder' | '/tool/stat' | '/tool/common-report' | '/tool/write-request' | '/tool/editbot' | '/tool/frontplay'
    label: string
  }

  const navItems: NavItem[] = [
    { to: '/tool/binder', label: '바인더' },
    { to: '/tool/stat', label: '통계' },
    { to: '/tool/common-report', label: '통용' },
    { to: '/tool/write-request', label: '작성요청' },
    { to: '/tool/editbot', label: '편집봇' },
    { to: '/tool/frontplay', label: 'FrontPlay' },
  ]

  function isActive(pathname: string, to: string) {
    return pathname === to || pathname.startsWith(`${to}/`)
  }

  let pathname = $derived(page.url.pathname)
  let { children }: { children?: Snippet } = $props()
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
        {#each navItems as item (item.to)}
          <tr>
            <td>
              <a
                href={resolve(item.to)}
                class={`block w-full rounded px-3 py-2 text-left transition z-text hover:no-underline hover:bg-gray-200 dark:hover:bg-gray-800 ${
                  isActive(pathname, item.to) ? 'bg-gray-200 font-semibold dark:bg-gray-800' : ''
                }`}
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
