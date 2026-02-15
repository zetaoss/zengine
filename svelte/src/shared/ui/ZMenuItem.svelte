<script lang="ts">
  import { createEventDispatcher } from 'svelte'

  const dispatch = createEventDispatcher<{ click: void }>()
</script>

<button type="button" class="w-full whitespace-nowrap px-3 py-2 text-left hover:bg-(--z-btn-hover)" on:click={() => dispatch('click')}>
  <slot />
</button>
