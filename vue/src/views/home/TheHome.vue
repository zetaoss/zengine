<script setup lang="ts">
import HomeAbout from './HomeAbout.vue'
import HomeComments from './HomeComments.vue'
import HomeForum from './HomeForum.vue'
import HomeNewPages from './HomeNewPages.vue'
import HomeOnelines from './HomeOnelines.vue'
</script>

<template>
  <div class="p-5">
    <h2 class="my-5 text-2xl font-bold">
      제타위키
    </h2>
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
          <a href="/onelines">한줄잡담</a>
        </header>
        <HomeOnelines />
      </section>
      <section>
        <header>
          <a href="/page-comments">문서댓글</a>
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
<style lang="scss" scoped>
section {
  @apply p-4 border rounded-md bg-z-card;

  header {
    @apply pb-3 mb-3 border-b;
  }

  a {
    @apply text-[1.1rem] font-bold text-z-text leading-5;
  }
}
</style>
