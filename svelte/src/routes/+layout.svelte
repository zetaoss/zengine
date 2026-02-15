<script lang="ts">
  import '../app.css'

  import TheNavbar from '$lib/components/navbar/TheNavbar.svelte'
  import CSiteFooter from '$shared/components/CSiteFooter.svelte'
  import CSiteRemocon from '$shared/components/CSiteRemocon.svelte'
  import ConfirmHost from '$shared/ui/confirm/ConfirmHost.svelte'
  import ToastHost from '$shared/ui/toast/ToastHost.svelte'
</script>

<ConfirmHost />
<ToastHost />
<div class="flex h-full min-h-screen flex-col">
  <TheNavbar />
  <main class="mx-auto w-full max-w-8xl flex-1">
    <slot />
  </main>
  <CSiteFooter />
</div>
<div class="fixed bottom-0 right-0 z-40">
  <CSiteRemocon />
</div>
