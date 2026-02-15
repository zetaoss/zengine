<script lang="ts">
  import './assets/forum-apex.css'

  import useAuthStore from '$lib/stores/auth'
  import ZButton from '$shared/ui/ZButton.svelte'

  import ForumPostList from './components/ForumPostList.svelte'

  const auth = useAuthStore()
  const canWrite = auth.canWrite
</script>

<div class="p-5">
  <h2 class="my-5 text-2xl font-bold">포럼</h2>

  <div class="mb-4 flex justify-end">
    <ZButton as="a" href="/forum/new" disabled={!$canWrite}>글쓰기</ZButton>
  </div>

  <ForumPostList />
</div>
