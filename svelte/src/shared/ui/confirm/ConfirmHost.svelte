<script lang="ts">
  import ZButton from '$shared/ui/ZButton.svelte'

  import { confirmState, handleConfirmCancel, handleConfirmOk } from './confirm'

  const colorClasses: Record<string, string> = {
    default: 'bg-[var(--z-btn-bg)]',
    danger: 'bg-[var(--z-danger-bg)]',
    ghost: 'bg-transparent',
    primary: 'bg-[var(--z-primary-bg)]',
  }
</script>

{#if $confirmState.show}
  <div class="fixed inset-0 z-50 flex items-center justify-center">
    <button class="absolute inset-0 bg-black/40" aria-label="Close" on:click={() => $confirmState.closable && handleConfirmCancel()}
    ></button>
    <div class="relative w-full max-w-md rounded-lg bg-white p-6 shadow-lg dark:bg-slate-900">
      <div class="text-sm text-slate-700 dark:text-slate-200">{$confirmState.message}</div>
      <div class="mt-6 flex justify-end gap-2">
        <ZButton color="ghost" class="px-4 py-2" onclick={handleConfirmCancel}>
          {$confirmState.cancelText}
        </ZButton>
        <ZButton color={$confirmState.okColor} class={`px-4 py-2 ${colorClasses[$confirmState.okColor] ?? ''}`} onclick={handleConfirmOk}>
          {$confirmState.okText}
        </ZButton>
      </div>
    </div>
  </div>
{/if}
