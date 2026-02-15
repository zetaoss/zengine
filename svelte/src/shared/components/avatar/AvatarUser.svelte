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
  href={showLink ? `/user/${user.name}` : undefined}
  class={`inline-flex items-center align-middle ${showLink ? 'rounded-lg text-gray-400 hover:no-underline hover:bg-gray-200 dark:hover:bg-gray-700' : ''}`}
>
  <AvatarIcon {user} {size} {showBorder} />
  {#if showName}
    <span class="px-1">{user.name}</span>
  {/if}
</svelte:element>
