<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <div class="h-full grid md:grid-cols-[120px_1fr]">
    <div class="z-card p-4">
      <p class="text-lg font-bold">도구</p>
      <RouterLink class="rl z-card-hover" to="/tool/common-report">통용</RouterLink>
      <RouterLink class="rl z-card-hover" to="/tool/write-request">작성요청</RouterLink>
      <RouterLink class="rl z-card-hover" to="/tool/frontbox">FrontBox</RouterLink>
    </div>
    <div>
      <RouterView />
    </div>
  </div>
</template>

<style scoped>
.rl {
  @apply block p-2;

  &:hover {
    @apply no-underline rounded;
  }
}
</style>
