<svelte:options customElement={{ tag: 'page-meta', shadow: 'none' }} />

<script lang="ts">
  import { mdiClockOutline } from '@mdi/js'

  import getRLCONF from '$lib/utils/rlconf'
  import AvatarIcon from '$shared/components/avatar/AvatarIcon.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  export let historyhref: string

  const { lastmod, contributors } = getRLCONF()

  const asDate = (v: string) => (v && v.length >= 8 ? `${v.substring(0, 4)}-${v.substring(4, 6)}-${v.substring(6, 8)}` : v)
</script>

{#if contributors?.length}
  <a href={historyhref} class="z-text2 inline-flex items-center gap-1">
    <ZIcon path={mdiClockOutline} class="h-4 w-4" />
    {asDate(lastmod)}
  </a>
  <span class="pl-3 -space-x-0.5">
    {#each contributors as u (u.id)}
      <a href={`/profile/${u.name}`}>
        <AvatarIcon user={u} showBorder={true} />
      </a>
    {/each}
  </span>
{/if}
