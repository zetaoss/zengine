<svelte:options customElement={{ tag: 'page-menu', shadow: 'none' }} />

<script lang="ts">
  import { mdiDotsVertical } from '@mdi/js'
  import { onMount } from 'svelte'

  import getLinks from '$lib/utils/getLinks'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  let root: HTMLDetailsElement | null = null

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

<details bind:this={root} class="page-menu relative print:hidden z-link">
  <summary class="page-btn cursor-pointer" aria-label="Page menu">
    <ZIcon path={mdiDotsVertical} />
  </summary>

  <div
    class="page-menu-panel absolute z-30 right-0 border rounded bg-white shadow-md dark:bg-neutral-800 text-sm text-black dark:text-white"
  >
    <ul class="page-menu-items">
      {#each links as l, i (i)}
        <!-- svelte-ignore a11y_accesskey -->
        <li><a href={l.href} accesskey={l.accesskey} title={l.title}>{l.text}</a></li>
      {/each}
    </ul>
  </div>
</details>

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
    text-decoration: none;
    background-color: #8883;
  }

  .page-menu-items {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .page-menu-items a {
    display: block;
    white-space: nowrap;
    padding: 0.25rem 1.5rem;
    color: var(--text);
  }

  .page-menu-items a:hover {
    text-decoration: none;
    background-color: #8883;
  }
</style>
