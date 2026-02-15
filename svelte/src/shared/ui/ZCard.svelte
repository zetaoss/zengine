<section class="z-base2 z-ring2 rounded-lg p-4 shadow-sm ring-1">
  {#if $$slots.header}
    <header class="mb-3 border-b pb-3 text-base font-semibold">
      <slot name="header" />
    </header>
  {/if}
  <slot />
</section>
