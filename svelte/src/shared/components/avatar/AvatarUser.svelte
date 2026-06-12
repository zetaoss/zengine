<script lang="ts">
  import type { User } from '$shared/types/user'

  import AvatarIcon from './AvatarIcon.svelte'

  export let user: User
  export let size = 18
  export let showName = true
  export let showLink = true
  export let showBorder = false
</script>

<svelte:element
  this={showLink ? 'a' : 'span'}
  href={showLink ? `/user/${encodeURIComponent(user.name.replace(/ /g, '_'))}` : undefined}
  class={`inline-flex items-center align-middle ${showLink ? 'rounded-lg text-a-gray-600/80 hover:no-underline hover:bg-a-gray-200' : ''}`}
>
  <AvatarIcon {user} {size} {showBorder} />
  {#if showName}
    <span class="px-1">{user.name}</span>
  {/if}
</svelte:element>
