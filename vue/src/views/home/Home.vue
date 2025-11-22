<!-- eslint-disable vue/multi-word-component-names -->
<script setup lang="ts">
import HomeAbout from './HomeAbout.vue'
import HomeComments from './HomeComments.vue'
import HomeForum from './HomeForum.vue'
import HomeNewPages from './HomeNewPages.vue'
import HomeOnelines from './HomeOnelines.vue'
</script>

<template>
  <div class="p-5">
    <h2 class="my-5 text-2xl font-bold">제타위키</h2>
    <div class="my-5 grid gap-4 md:grid-cols-2">
      <section>
        <HomeAbout />
      </section>
      <section>
        <header>
          <a href="/forum">포럼</a>
        </header>
        <HomeForum />
      </section>
    </div>
    <div class="my-5 grid gap-4 md:grid-cols-[2fr_2fr_1fr]">
      <section>
        <header>
          한줄잡담
        </header>
        <HomeOnelines />
      </section>
      <section>
        <header>
          문서댓글
        </header>
        <HomeComments />
      </section>
      <section>
        <header>
          <a href="/wiki/특수:새문서">새 문서</a>
        </header>
        <HomeNewPages />
      </section>
    </div>
  </div>
</template>

<style scoped>
section {
  background-color: var(--card-bg);
  box-shadow: 0 0 0 1px var(--card-border), 0 1px 2px 0 rgba(0, 0, 0, .04);
  border-radius: 8px;

  padding: 1rem;

  header {
    border-bottom-width: 1px;
    font-size: 1.1rem;
    font-weight: 600;
    margin-bottom: 0.75rem;
    padding-bottom: 0.75rem;
  }
}
</style>
