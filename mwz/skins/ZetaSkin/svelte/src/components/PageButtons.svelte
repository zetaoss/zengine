<svelte:options customElement={{ tag: 'page-buttons', shadow: 'none' }} />

<script lang="ts">
  import getLinks from '$lib/utils/getLinks'
  import getRLCONF from '$lib/utils/rlconf'

  const { wgAction } = getRLCONF()

  const links = getLinks(
    wgAction != 'view' && 'views.view',
    wgAction != 'edit' && ['views.edit', { accesskey: 'e' }],
    ['actions.viewsource', { accesskey: 'e' }],
    ['toolbox.whatlinkshere', { accesskey: 'j', text: '역링크' }],
    ['namespaces.talk', { accesskey: 't' }],
  )
</script>

{#each links as l, i (i)}
  <!-- svelte-ignore a11y_accesskey -->
  <a class="page-btn" href={l.href} id={l.id} title={l.title} accesskey={l.accesskey}>{l.text}</a>
{/each}
