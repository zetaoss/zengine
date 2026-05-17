<svelte:options customElement={{ tag: 'page-menu', shadow: 'none' }} />

<script lang="ts">
  import { mdiDotsVertical } from '@mdi/js'
  import { onMount } from 'svelte'

  import getLinks from '$lib/utils/getLinks'
  import getRLCONF from '$lib/utils/rlconf'
  import EditBotModal from '$shared/components/editbot/EditBotModal.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  let root: HTMLDetailsElement | null = null
  let showEditBotModal = $state(false)

  const { wgArticleId, wgUserGroups } = getRLCONF()
  const isSysop = (wgUserGroups || []).includes('sysop')

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

  function getCurrentTitle() {
    const heading = document.getElementById('firstHeading')?.textContent?.trim()
    if (heading) return heading

    const path = decodeURIComponent(window.location.pathname)
    return path.split('/').filter(Boolean).pop()?.replaceAll('_', ' ') || '현재'
  }

  function requestEditTask() {
    if (!isSysop || wgArticleId < 1) return
    showEditBotModal = true
    close()
  }

  function closeEditBotModal() {
    showEditBotModal = false
  }

  let editBotTarget = $derived.by(() => {
    if (!showEditBotModal || wgArticleId < 1) return null
    return {
      title: getCurrentTitle(),
      storeUrl: '/api/editbot/from-page',
      requestType: 'edit' as const,
      pageId: wgArticleId,
    }
  })

  function onEditBotCreated() {
    showEditBotModal = false
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
      {#if isSysop}
        <li>
          <button type="button" onclick={requestEditTask}> 편집 요청 </button>
        </li>
      {/if}
    </ul>
  </div>
</details>

<EditBotModal show={showEditBotModal} target={editBotTarget} onClose={closeEditBotModal} onCreated={onEditBotCreated} />

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
    border: 0;
    background: transparent;
    cursor: pointer;
    font: inherit;
  }

  .page-menu-items button:disabled {
    cursor: not-allowed;
    opacity: 0.5;
  }

  .page-menu-items a:hover,
  .page-menu-items button:hover {
    background-color: #8883;
    text-decoration: none;
  }
</style>
