<script lang="ts">
  import type { Snippet } from 'svelte'

  import CButton from './CButton.svelte'

  let { onclick, children }: { onclick?: () => void; children?: Snippet } = $props()
</script>

<CButton
  variant="ghost"
  class="w-full !justify-start !rounded-none px-3 py-2 !font-normal whitespace-nowrap gap-2"
  {onclick}
>
  {@render children?.()}
</CButton>
