<svelte:options runes={true} />

<script lang="ts">
  import { mdiAccount } from '@mdi/js'
  import { onMount } from 'svelte'

  import { page } from '$app/state'
  import useAuthStore from '$lib/stores/auth'
  import AvatarIcon from '$shared/components/avatar/AvatarIcon.svelte'
  import { useDismissable } from '$shared/composables/useDismissable'
  import type { Link } from '$shared/types/links'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import getShortcut from '$shared/utils/shortcut'

  const auth = useAuthStore()
  const { userInfo, isLoggedIn } = auth

  let root: HTMLElement | null = null
  let open = $state(false)
  let lastPath = ''

  const toggle = () => (open = !open)
  const close = () => (open = false)

  useDismissable(() => root, {
    enabled: () => open,
    onDismiss: close,
  })

  $effect(() => {
    const pathname = page.url.pathname
    if (pathname !== lastPath) {
      lastPath = pathname
      close()
    }
  })

  const links: Link[] = $derived(
    ($isLoggedIn
      ? [
          { text: '프로필', href: `/user/${$userInfo?.name ?? ''}` },
          { text: '사용자 문서', href: '/wiki/특수:내사용자문서', accesskey: '.' },
          { text: '사용자 토론', href: '/wiki/특수:내사용자토론', accesskey: 'n' },
          { text: '환경 설정', href: '/wiki/특수:환경설정' },
          { text: '주시문서 목록', href: '/wiki/특수:주시문서목록', accesskey: 'l' },
          { text: '기여', href: '/wiki/특수:내기여', accesskey: 'y' },
          { text: '업로드', href: '/wiki/특수:올리기' },
          { text: '특수문서', href: '/wiki/특수:특수문서' },
          { text: '로그아웃', href: '/logout' },
        ]
      : [
          { text: '로그인', href: '/login', accesskey: 'o' },
          { text: '계정 만들기', href: '/wiki/특수:계정만들기' },
        ]
    ).map((link) => {
      const shortcut = getShortcut(link.accesskey)
      return {
        ...link,
        title: shortcut ? `${link.text} (${shortcut})` : link.text,
      }
    }),
  )

  onMount(() => {
    auth.update()
  })
</script>

<div bind:this={root} class="md:group order-2 ml-auto contents md:relative md:inline-block">
  <button
    type="button"
    class={`cursor-pointer order-2 ml-auto flex h-12 w-12 items-center justify-center hover:bg-gray-800 md:w-auto md:px-3 ${open ? 'bg-gray-800' : ''}`}
    aria-expanded={open}
    onclick={toggle}
  >
    {#if $isLoggedIn && $userInfo}
      <AvatarIcon user={$userInfo} size={20} />
    {:else}
      <ZIcon size={20} path={mdiAccount} />
    {/if}
  </button>

  <div
    class={`order-3 z-40 bg-gray-800 md:absolute md:right-0 md:m-1 md:rounded md:border w-full md:w-auto ${open ? 'block' : 'hidden'}`}
    role="menu"
    tabindex="-1"
    onclick={(e) => e.stopPropagation()}
    onkeydown={(e) => e.stopPropagation()}
  >
    <nav class="grid w-full grid-cols-3 py-1 md:block md:w-fit md:whitespace-nowrap">
      {#each links as l, i (i)}
        <!-- svelte-ignore a11y_accesskey -->
        <a
          class="block p-2 px-8 text-xs text-white hover:bg-gray-700 hover:no-underline"
          href={l.href}
          title={l.title}
          accesskey={l.accesskey}
          rel="external"
          data-sveltekit-reload
          onclick={close}
        >
          {l.text}
        </a>
      {/each}
    </nav>
  </div>
</div>
