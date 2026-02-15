<script lang="ts">
  import { resolve } from '$app/paths'
  import ZCard from '$shared/ui/ZCard.svelte'

  import HomeAbout from './HomeAbout.svelte'
  import HomeComments from './HomeComments.svelte'
  import HomeForum from './HomeForum.svelte'
  import HomeNewPages from './HomeNewPages.svelte'
  import HomeOnelines from './HomeOnelines.svelte'
</script>

<div class="p-5">
  <h2 class="my-5 text-2xl font-bold">제타위키</h2>
  <div class="my-5 grid gap-4 md:grid-cols-2">
    <ZCard>
      <HomeAbout />
    </ZCard>
    <ZCard>
      <svelte:fragment slot="header">
        <a href={resolve('/forum')}>포럼</a>
      </svelte:fragment>
      <HomeForum />
    </ZCard>
  </div>
  <div class="my-5 grid gap-4 md:grid-cols-[2fr_2fr_1fr]">
    <ZCard>
      <svelte:fragment slot="header">
        <a href={resolve('/onelines')}>한줄잡담</a>
      </svelte:fragment>
      <HomeOnelines />
    </ZCard>
    <ZCard>
      <svelte:fragment slot="header">문서댓글</svelte:fragment>
      <HomeComments />
    </ZCard>
    <ZCard>
      <svelte:fragment slot="header">
        <a href="/wiki/특수:새문서" rel="external" data-sveltekit-reload>새 문서</a>
      </svelte:fragment>
      <HomeNewPages />
    </ZCard>
  </div>
</div>
