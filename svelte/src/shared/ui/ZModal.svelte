<script lang="ts">
  import { mdiClose } from '@mdi/js'
  import { onMount } from 'svelte'
  import { createEventDispatcher } from 'svelte'

  import ZButton from '$shared/ui/ZButton.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  export let show = false
  export let title = ''
  export let okText = '확인'
  export let okColor: 'ghost' | 'default' | 'danger' | 'primary' = 'danger'
  export let okDisabled = false
  export let cancelText = '취소'
  export let closable = true
  export let backdropClosable = true

  const dispatch = createEventDispatcher<{ ok: void; cancel: void }>()

  onMount(() => {
    function onKeyup(e: KeyboardEvent) {
      if (e.key === 'Escape' && show && closable) dispatch('cancel')
    }

    document.addEventListener('keyup', onKeyup)
    return () => document.removeEventListener('keyup', onKeyup)
  })
</script>

{#if show}
  <div class="fixed inset-0 z-40 flex items-center justify-center bg-black/40">
    {#if backdropClosable}
      <button type="button" class="absolute inset-0 cursor-default bg-transparent" aria-label="닫기" onclick={() => dispatch('cancel')}
      ></button>
    {/if}

    <div role="dialog" aria-modal="true" class="relative w-full max-w-[60vw] rounded-md border bg-white dark:bg-gray-900 md:max-w-[40vw]">
      {#if closable}
        <ZButton color="ghost" class="float-right m-1" onclick={() => dispatch('cancel')}>
          <ZIcon path={mdiClose} />
        </ZButton>
      {/if}

      {#if title}
        <header class="border-b px-5 py-3">
          <h2 class="m-0 text-base font-semibold">{title}</h2>
        </header>
      {/if}

      <section class="p-5">
        <slot />
      </section>

      <footer class="flex justify-center gap-3 border-t px-4 py-3">
        <ZButton color={okColor} disabled={okDisabled} onclick={() => !okDisabled && dispatch('ok')}>
          {okText}
        </ZButton>
        <ZButton onclick={() => dispatch('cancel')}>{cancelText}</ZButton>
      </footer>
    </div>
  </div>
{/if}
