<!-- eslint-disable vue/multi-word-component-names -->
<template>
  <div class="h-full grid md:grid-cols-[120px_1fr]">
    <div class="bg-z-card p-4">
      <p class="text-lg font-bold">도구</p>
      <RouterLink class="rl" to="/tool/common-report">통용</RouterLink>
      <RouterLink class="rl" to="/tool/write-request">작성요청</RouterLink>
      <RouterLink class="rl" to="/tool/random">랜덤</RouterLink>
      <RouterLink class="rl" to="/tool/svg-to-png">svg-png</RouterLink>
      <RouterLink class="rl" to="/tool/frontbox">FrontBox</RouterLink>
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
    @apply no-underline bg-z-card-hover rounded;
  }
}
</style>
