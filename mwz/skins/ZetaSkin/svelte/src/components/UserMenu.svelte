<svelte:options customElement={{ tag: 'user-menu', shadow: 'none' }} />

<script lang="ts">
  import { mdiAccount } from '@mdi/js'
  import { onMount } from 'svelte'

  import getRLCONF from '$lib/utils/rlconf'
  import AvatarIcon from '$shared/components/avatar/AvatarIcon.svelte'
  import type Item from '$shared/components/navbar/items'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  export let userMenu: Record<string, Item> | Item[] | string = {}

  const { wgUserId, wgUserName } = getRLCONF()

  let root: HTMLElement | null = null
  let open = false
  const parseMenu = (input: string) => {
    if (!input) return {}
    try {
      return JSON.parse(input) as Record<string, Item>
    } catch {
      return {}
    }
  }

  const getConfigMenu = () => {
    try {
      const mw = (
        window as Window & {
          mw?: { config?: { get?: (key: string) => unknown } }
        }
      ).mw
      const val = mw?.config?.get?.('myMenu')
      return (val && typeof val === 'object' ? val : {}) as Record<string, Item>
    } catch {
      return {}
    }
  }

  const toggle = () => {
    open = !open
  }

  const close = () => {
    open = false
  }

  const onMouseDown = (e: MouseEvent) => {
    if (!open) return
    if (!root) return
    if (root.contains(e.target as Node)) return
    close()
  }

  const onKeyDown = (e: KeyboardEvent) => {
    if (!open) return
    if (e.key === 'Escape') close()
  }

  onMount(() => {
    document.addEventListener('mousedown', onMouseDown)
    document.addEventListener('keydown', onKeyDown)
    return () => {
      document.removeEventListener('mousedown', onMouseDown)
      document.removeEventListener('keydown', onKeyDown)
    }
  })

  $: parsedMenu = typeof userMenu === 'string' ? parseMenu(userMenu) : userMenu && Object.keys(userMenu).length ? userMenu : getConfigMenu()

  $: items = Object.values(parsedMenu)
</script>

<div bind:this={root} class="md:group order-2 ml-auto contents md:relative md:inline-block">
  <button
    type="button"
    class="order-2 ml-auto flex h-12 cursor-pointer items-center px-3 hover:bg-gray-800"
    class:bg-gray-800={open}
    aria-expanded={open}
    on:click={toggle}
  >
    {#if wgUserId}
      <AvatarIcon size={20} user={{ id: wgUserId, name: wgUserName }} />
    {:else}
      <ZIcon size={20} path={mdiAccount} />
    {/if}
  </button>

  <div
    class={`order-3 z-40 bg-gray-800 md:absolute md:right-0 md:m-1 md:rounded md:group-hover:block md:border w-full md:w-auto ${open ? 'block' : 'hidden'}`}
  >
    <nav class="grid grid-cols-3 w-full py-1 md:w-fit md:block md:whitespace-nowrap">
      {#each items as item (item.text)}
        <!-- svelte-ignore a11y_accesskey -->
        <a
          href={item.href}
          title={(item as Item & { title?: string }).title}
          accesskey={(item as Item & { accesskey?: string }).accesskey}
          class="block p-2 px-8 text-xs text-white hover:bg-gray-700 hover:no-underline"
          on:click={close}
        >
          {item.text}
        </a>
      {/each}
    </nav>
  </div>
</div>

<style>
  :global(user-menu) {
    display: contents;
  }
</style>
