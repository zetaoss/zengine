<script lang="ts">
  import { mdiClose } from '@mdi/js'
  import { onMount } from 'svelte'
  import { createEventDispatcher } from 'svelte'

  import ZButton from '$shared/ui/ZButton.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  export let show = false
  export let title = ''
  export let titleIconPath: string | undefined = undefined
  export let okText = '확인'
  export let okColor: 'ghost' | 'default' | 'danger' | 'primary' = 'danger'
  export let okDisabled = false
  export let cancelText = '취소'
  export let closable = true
  export let backdropClosable = true
  export let panelClass = 'max-w-[60vw] md:max-w-[40vw]'

  const dispatch = createEventDispatcher<{ ok: void; cancel: void }>()

  onMount(() => {
    function onKeyup(e: KeyboardEvent) {
      if (e.key === 'Escape' && show && closable) dispatch('cancel')
    }

    document.addEventListener('keyup', onKeyup)
    return () => {
      document.removeEventListener('keyup', onKeyup)
    }
  })
</script>

{#if show}
  <div class="fixed inset-0 z-40 flex items-center justify-center bg-black/40">
    {#if backdropClosable}
      <button type="button" class="absolute inset-0 cursor-default bg-transparent" aria-label="닫기" onclick={() => dispatch('cancel')}
      ></button>
    {/if}

    <div
      role="dialog"
      aria-modal="true"
      class={`relative flex max-h-[calc(100dvh-2rem)] w-full min-h-0 flex-col overflow-hidden rounded-md border bg-white dark:bg-gray-900 ${panelClass}`}
    >
      {#if title}
        <header class="flex min-h-12 items-stretch justify-between gap-2 border-b pl-5">
          <div role="heading" aria-level="2" class="m-0 flex min-w-0 items-center gap-2 py-3 text-base font-semibold leading-tight">
            {#if titleIconPath}
              <ZIcon path={titleIconPath} />
            {/if}
            <span>{title}</span>
          </div>
          {#if closable}
            <ZButton color="ghost" class="self-stretch rounded-none px-4 py-0" onclick={() => dispatch('cancel')}>
              <ZIcon path={mdiClose} />
            </ZButton>
          {/if}
        </header>
      {:else if closable}
        <div class="flex justify-end border-b px-3 py-2">
          <ZButton color="ghost" onclick={() => dispatch('cancel')}>
            <ZIcon path={mdiClose} />
          </ZButton>
        </div>
      {/if}

      <section class="min-h-0 flex-1 overflow-y-auto p-5">
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
