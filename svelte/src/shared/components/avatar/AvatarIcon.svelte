<script lang="ts">
  import { getAvatarBaseUrl } from '$shared/config'
  import type { User } from '$shared/types/user'

  type AvatarType = 0 | 1 | 2 | 3

  export let user: User | null = null
  export let size = 24
  export let typ: AvatarType = 0
  export let showBorder = false

  const baseUrl = getAvatarBaseUrl()

  $: src = (() => {
    if (!user) return ''

    const u = new URL(`${baseUrl}/u/${user.id}`)
    u.searchParams.set('s', String(size))
    if (typ) u.searchParams.set('t', String(typ))

    const v = localStorage.getItem('v')
    if (v) u.searchParams.set('v', v)

    return u.toString()
  })()
</script>

<span
  class={`inline-flex items-center justify-center overflow-hidden rounded-full box-border align-middle relative hover:z-40 ${
    showBorder ? 'ring-2 ring-white dark:ring-gray-900 outline -outline-offset-1 outline-black/5 dark:outline-white/10' : ''
  }`}
  style={`height: ${size}px; width: ${size}px; background: #f0f0f0;`}
  title={user?.name}
>
  {#if user?.id === 0}
    <span class="flex h-full w-full items-center justify-center text-xs text-gray-500">?</span>
  {:else if user}
    <img
      class="h-full w-full"
      {src}
      width={size}
      height={size}
      loading="lazy"
      decoding="async"
      referrerpolicy="no-referrer"
      alt={user.name}
    />
  {/if}
</span>
