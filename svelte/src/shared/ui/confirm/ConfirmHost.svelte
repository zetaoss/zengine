<svelte:options customElement={{ tag: 'confirm-host', shadow: 'none' }} />

<script lang="ts">
  import CButton from '$shared/ui/CButton.svelte'

  import { confirmState, handleConfirmCancel, handleConfirmOk } from './confirm'
</script>

{#if $confirmState.show}
  <div class="fixed inset-0 z-50 flex items-center justify-center">
    <button class="absolute inset-0 bg-black/40" aria-label="Close" on:click={() => $confirmState.closable && handleConfirmCancel()}
    ></button>
    <div class="relative w-full max-w-md rounded-lg bg-(--background-color-base) p-6 shadow-lg border">
      <div class="text-sm text-slate-700">{$confirmState.message}</div>
      <div class="mt-6 flex justify-end gap-2">
        <CButton variant="ghost" class="px-4 py-2" onclick={handleConfirmCancel}>
          {$confirmState.cancelText}
        </CButton>
        <CButton
          variant={$confirmState.okVariant}
          class={`px-4 py-2 confirm-btn-${$confirmState.okVariant}`}
          onclick={handleConfirmOk}
        >
          {$confirmState.okText}
        </CButton>
      </div>
    </div>
  </div>
{/if}

<style>
  :global(.confirm-btn-outline) {
    background-color: var(--background-color-interactive);
  }
  :global(.confirm-btn-default) {
    background-color: var(--color-progressive);
    color: white;
  }
  :global(.confirm-btn-ghost) {
    background-color: transparent;
  }
  :global(.confirm-btn-destructive) {
    background-color: var(--color-destructive);
    color: white;
  }
</style>
