<!-- eslint-disable vue/multi-word-component-names -->
<script setup lang="ts">
import ZCard from '@common/ui/ZCard.vue'
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
      <ZCard>
        <HomeAbout />
      </ZCard>
      <ZCard>
        <template #header>
          <a href="/forum">포럼</a>
        </template>
        <HomeForum />
      </ZCard>
    </div>
    <div class="my-5 grid gap-4 md:grid-cols-[2fr_2fr_1fr]">
      <ZCard>
        <template #header>
          한줄잡담
        </template>
        <HomeOnelines />
      </ZCard>
      <ZCard>
        <template #header>
          문서댓글
        </template>
        <HomeComments />
      </ZCard>
      <ZCard>
        <template #header>
          <a href="/wiki/특수:새문서">새 문서</a>
        </template>
        <HomeNewPages />
      </ZCard>
    </div>
  </div>
</template>
