<svelte:options customElement={{ tag: 'page-meta', shadow: 'none' }} />

<script lang="ts">
  import { mdiClockOutline } from '@mdi/js'

  import getRLCONF from '$lib/utils/rlconf'
  import AvatarIcon from '$shared/components/avatar/AvatarIcon.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  const { contributors, mm, revtime } = getRLCONF()
  const historyhref = mm['views']['history']['href'] ?? null
  const revtimedate = `${revtime.substring(0, 4)}-${revtime.substring(4, 6)}-${revtime.substring(6, 8)}`
</script>

{#if contributors?.length}
  <a href={historyhref} class="z-text2 inline-flex items-center gap-1">
    <ZIcon path={mdiClockOutline} class="h-4 w-4" />
    {revtimedate}
  </a>
  <span class="pl-3 -space-x-0.5">
    {#each contributors as u (u.id)}
      <a href={`/user/${u.name}`}>
        <AvatarIcon user={u} showBorder={true} />
      </a>
    {/each}
  </span>
{/if}
