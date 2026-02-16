<script lang="ts">
  import './assets/forum-apex.css'

  import RouteLinkButton from '$lib/components/RouteLinkButton.svelte'
  import useAuthStore from '$lib/stores/auth'

  import ForumPostList from './components/ForumPostList.svelte'

  const auth = useAuthStore()
  const canWrite = auth.canWrite
</script>

<div class="p-5">
  <h2 class="my-5 text-2xl font-bold">포럼</h2>

  <div class="mb-4 flex justify-end">
    <RouteLinkButton to="/forum/new" disabled={!$canWrite}>글쓰기</RouteLinkButton>
  </div>

  <ForumPostList />
</div>
