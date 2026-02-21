<svelte:options customElement={{ tag: 'page-meta', shadow: 'none' }} />

<script lang="ts">
  import { mdiClockOutline } from '@mdi/js'

  import getRLCONF from '$lib/utils/rlconf'
  import AvatarIcon from '$shared/components/avatar/AvatarIcon.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  const { contributors, menu, lastModified } = getRLCONF()
  const historyHref = menu?.views?.history?.href ?? null
  const lastMod = `${lastModified.substring(0, 4)}-${lastModified.substring(4, 6)}-${lastModified.substring(6, 8)}`
</script>

{#if contributors?.length}
  {#if historyHref}
    <a href={historyHref} class="z-text2 inline-flex items-center gap-1">
      <ZIcon path={mdiClockOutline} class="h-4 w-4" />
      {lastMod}
    </a>
  {/if}
  <span class="pl-3 -space-x-0.5">
    {#each contributors as u (u.id)}
      <a href={`/user/${u.name}`}>
        <AvatarIcon user={u} showBorder={true} />
      </a>
    {/each}
  </span>
{/if}
