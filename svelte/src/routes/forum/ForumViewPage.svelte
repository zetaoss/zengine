<svelte:options runes={true} />

<script lang="ts">
  import './assets/forum-apex.css'

  import { page } from '$app/state'
  import ZButton from '$shared/ui/ZButton.svelte'

  import ForumPostList from './components/ForumPostList.svelte'
  import ViewerApex from './viewer/ViewerApex.svelte'

  let postId = $derived(Number(page.params.id) || 0)
  let pageNumber = $derived.by(() => {
    const p = Number(page.url.searchParams.get('page'))
    return Number.isFinite(p) && p > 0 ? p : 1
  })
</script>

<div class="p-5">
  <h2 class="my-5 text-2xl font-bold">포럼</h2>

  <div class="flex justify-end py-2">
    <ZButton as="a" href={`/forum${pageNumber === 1 ? '' : `?page=${pageNumber}`}`}>목록</ZButton>
  </div>

  <ViewerApex {postId} />

  <div class="mt-10">
    <ForumPostList currentPostId={postId} title={pageNumber === 1 ? '글 목록 (1페이지)' : `글 목록 (page ${pageNumber})`} />
  </div>
</div>
