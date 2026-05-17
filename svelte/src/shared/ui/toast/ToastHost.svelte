<svelte:options customElement={{ tag: 'toast-host', shadow: 'none' }} />

<script lang="ts">
  import { fly } from 'svelte/transition'

  import ZButtonLink from '../ZButtonLink.svelte'
  import { dismissToast, toasts } from './toast'
</script>

<div class="fixed right-2 top-2 z-50">
  {#each $toasts as toast (toast.id)}
    {#if toast.placement === 'top-right'}
      <div
        class="mt-1 w-56 rounded bg-slate-500/80 text-white shadow-md"
        in:fly={{ y: 6, duration: 200 }}
        out:fly={{ y: 6, duration: 200 }}
      >
        <div class="flex min-h-10 items-center px-4 py-2 leading-tight">{toast.message}</div>
        <hr class="toast-bar m-0 h-1 rounded border-0 bg-black/50" style={`animation-duration: ${toast.timeout}ms`} />
      </div>
    {/if}
  {/each}
</div>

<div class="pointer-events-none fixed inset-0 z-50 flex items-center justify-center p-4">
  {#each $toasts as toast (toast.id)}
    {#if toast.placement === 'center'}
      <div
        class="pointer-events-auto w-full max-w-sm rounded border bg-white text-(--z-text) shadow-lg dark:bg-gray-900"
        in:fly={{ y: 8, duration: 200 }}
        out:fly={{ y: 8, duration: 200 }}
      >
        <div class="flex min-h-14 items-center justify-center px-5 py-4 text-center leading-tight">{toast.message}</div>
        {#if toast.action}
          <div class="flex justify-center px-5 pb-4">
            <ZButtonLink href={toast.action.href} color="primary" onclick={() => dismissToast(toast.id)}>
              {toast.action.label}
            </ZButtonLink>
          </div>
        {/if}
        <hr class="toast-bar m-0 h-1 rounded border-0 bg-black/20 dark:bg-white/20" style={`animation-duration: ${toast.timeout}ms`} />
      </div>
    {/if}
  {/each}
</div>

<style>
  @keyframes toast-bar {
    from {
      width: 100%;
    }

    to {
      width: 0%;
    }
  }

  .toast-bar {
    animation: toast-bar linear forwards;
  }
</style>
