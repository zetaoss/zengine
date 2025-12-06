<!-- ZCard.vue -->
<template>
  <section class="rounded-lg ring-1 ring-[var(--z-card-border)] bg-[var(--z-card-bg)] shadow-sm p-4">
    <header v-if="$slots.header" class="mb-3 pb-3 border-b text-base font-semibold">
      <slot name="header" />
    </header>
    <slot />
  </section>
</template>
