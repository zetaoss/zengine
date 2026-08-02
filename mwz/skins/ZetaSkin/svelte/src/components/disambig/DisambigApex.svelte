<svelte:options customElement={{ tag: 'disambig-apex', shadow: 'none' }} />

<script lang="ts">
  import getRLCONF from '$lib/utils/rlconf'

  import DisambigModal from './DisambigModal.svelte'

  const { disambig, disambigRegistration, wgArticleId } = getRLCONF()
  let showModal = false
</script>

{#if disambig?.nodes?.length && disambigRegistration}
  <aside
    class={`my-3 w-fit max-w-full text-[0.8rem] ${disambig?.nodes?.length ? 'rounded-md p-2 pl-5' : ''}`}
    aria-label={`${disambigRegistration.baseTitle} 동음이의 문서`}
  >
    <div class="flex items-start gap-1">
      <button
        type="button"
        class="inline-flex flex-none rounded p-0.5 hover:bg-muted"
        title={`${disambigRegistration.targetTitle} ${disambigRegistration.exists ? '편집' : '생성'}`}
        aria-label={`${disambigRegistration.targetTitle} ${disambigRegistration.exists ? '편집' : '생성'}`}
        onclick={() => (showModal = true)}
      >
        <img class="size-5" src="https://upload.wikimedia.org/wikipedia/commons/4/4a/Disambig_grey.svg" alt="" aria-hidden="true" />
      </button>

      {#if disambig?.nodes?.length}
        <ul class="m-0 flex min-w-0 flex-1 list-none flex-col gap-0.5 p-0">
          {#each disambig.nodes as node, index (`${node.title}:${index}`)}
            <li class="w-full min-w-0">
              <div class="inline-flex min-w-0 flex-1 flex-wrap items-baseline gap-x-2 px-1.5 py-0.5">
                다른 뜻에 대해서는
                {#if node.id === wgArticleId}
                  <span class="font-semibold" aria-current="page">{node.text}</span>
                {:else}
                  <a class:new={Boolean(node.new)} href={node.href}>{node.text}</a>
                {/if}
                {#if node.description}
                  <span class="text-[0.6rem] text-muted-foreground">{node.description}</span>
                {/if}
                문서를 참조하십시오.
              </div>
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  </aside>

  {#if showModal}
    <DisambigModal
      show
      baseTitle={disambigRegistration.baseTitle}
      targetExists={disambigRegistration.exists}
      sourceTitle={disambigRegistration.sourceTitle}
      targetTitle={disambigRegistration.targetTitle}
      onClose={() => (showModal = false)}
    />
  {/if}
{/if}

<style>
  :global(disambig-apex) {
    display: block;
  }

  a.new {
    color: var(--color-destructive);
  }
</style>
