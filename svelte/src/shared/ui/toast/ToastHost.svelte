<script lang="ts">
  import { fly } from 'svelte/transition'

  import { toasts } from './toast'
</script>

<div class="fixed right-2 top-2 z-50">
  {#each $toasts as toast (toast.id)}
    <div class="mt-1 w-56 rounded bg-slate-500/80 text-white shadow-md" in:fly={{ y: 6, duration: 200 }} out:fly={{ y: 6, duration: 200 }}>
      <div class="px-4 py-2">{toast.message}</div>
      <hr class="toast-bar h-1 rounded bg-black/50" style={`animation-duration: ${toast.timeout}ms`} />
    </div>
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
