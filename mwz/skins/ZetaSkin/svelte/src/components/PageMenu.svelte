<svelte:options customElement={{ tag: 'page-menu', shadow: 'none' }} />

<script lang="ts">
  import { mdiDotsVertical } from '@mdi/js'
  import { onMount } from 'svelte'

  import getLinks from '$lib/utils/getLinks'
  import getRLCONF from '$lib/utils/rlconf'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  import DisambigModal from './disambig/DisambigModal.svelte'

  let root: HTMLDetailsElement | null = null
  let showDisambigModal = false

  const { disambig, disambigRegistration } = getRLCONF()
  const links = getLinks(
    ['views.history', { accesskey: 'h' }],
    ['actions.delete', { accesskey: 'd' }],
    ['actions.move', { accesskey: 'm' }],
    ['actions.protect', { accesskey: '=' }],
    ['actions.unprotect', { accesskey: '=' }],
    ['views.watch', { accesskey: 'w' }],
    ['views.unwatch', { accesskey: 'w' }],
    ['toolbox.print', { accesskey: 'p' }],
    'toolbox.permalink',
    'toolbox.info',
  )

  const close = () => {
    if (!root) return
    root.open = false
  }

  const openDisambigModal = () => {
    close()
    showDisambigModal = true
  }

  const onMouseDown = (e: MouseEvent) => {
    if (!root?.open) return
    if (root.contains(e.target as Node)) return
    close()
  }

  const onKeyDown = (e: KeyboardEvent) => {
    if (!root?.open) return
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
</script>

<details bind:this={root} class="page-menu relative print:hidden text-a-sky-600">
  <summary class="page-btn cursor-pointer" aria-label="Page menu">
    <ZIcon path={mdiDotsVertical} />
  </summary>

  <div class="page-menu-panel absolute z-30 right-0 border rounded bg-background shadow-md text-sm text-foreground">
    <ul class="page-menu-items">
      {#each links as l, i (i)}
        <!-- svelte-ignore a11y_accesskey -->
        <li><a href={l.href} accesskey={l.accesskey} title={l.title}>{l.text}</a></li>
      {/each}
      {#if disambigRegistration && !disambig?.nodes?.length}
        <li>
          <button type="button" onclick={openDisambigModal}>동음이의 {disambigRegistration.exists ? '등록' : '생성'}</button>
        </li>
      {/if}
    </ul>
  </div>
</details>

{#if disambigRegistration && !disambig?.nodes?.length && showDisambigModal}
  <DisambigModal
    show
    baseTitle={disambigRegistration.baseTitle}
    targetExists={disambigRegistration.exists}
    sourceTitle={disambigRegistration.sourceTitle}
    targetTitle={disambigRegistration.targetTitle}
    onClose={() => (showDisambigModal = false)}
  />
{/if}

<style>
  .page-menu summary {
    list-style: none;
  }

  .page-menu summary::-webkit-details-marker {
    display: none;
  }

  .page-menu summary::marker {
    content: '';
  }

  .page-menu[open] > .page-btn {
    background-color: #8883;
    text-decoration: none;
  }

  .page-menu-items {
    padding: 0;
    margin: 0;
    list-style: none;
  }

  .page-menu-items a,
  .page-menu-items button {
    display: block;
    width: 100%;
    padding: 0.25rem 1.5rem;
    color: var(--text);
    text-align: left;
    white-space: nowrap;
  }

  .page-menu-items button {
    cursor: pointer;
  }

  .page-menu-items a:hover,
  .page-menu-items button:hover {
    background-color: #8883;
    text-decoration: none;
  }
</style>
