<script lang="ts">
  import { goto } from '$app/navigation'
  import { resolve } from '$app/paths'
  import useAuthStore from '$lib/stores/auth'
  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZModal from '$shared/ui/ZModal.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'

  import type { Post } from '../types'
  import ViewerHTML from './ViewerHTML.svelte'
  import ViewerReplies from './ViewerReplies.svelte'

  export let postId: number

  const auth = useAuthStore()
  const { canEdit, canDelete, canWrite } = auth

  let post: Post | null = null
  let showModal = false
  let isLoading = false

  async function fetchData() {
    if (!postId) return
    isLoading = true
    try {
      const [data, err] = await httpy.get<Post>(`/api/posts/${postId}`)
      if (err) {
        console.error('Failed to load post:', err)
        post = null
        return
      }
      post = data
    } finally {
      isLoading = false
    }
  }

  function edit() {
    goto(resolve(`/forum/${postId}/edit`))
  }

  function del() {
    showModal = true
  }

  async function modalOK() {
    showModal = false
    const [, err] = await httpy.delete(`/api/posts/${postId}`)
    if (err) {
      console.error(err)
      return
    }
    window.location.href = '/forum'
  }

  $: if (postId) fetchData()

  function formatDate(date: string) {
    const d = new Date(date)
    if (Number.isNaN(d.getTime())) return date
    const yyyy = d.getFullYear()
    const mm = String(d.getMonth() + 1).padStart(2, '0')
    const dd = String(d.getDate()).padStart(2, '0')
    const hh = String(d.getHours()).padStart(2, '0')
    const mi = String(d.getMinutes()).padStart(2, '0')
    return `${yyyy}-${mm}-${dd} ${hh}:${mi}`
  }
</script>

<ZModal show={showModal} on:ok={modalOK} on:cancel={() => (showModal = false)}>글을 삭제하시겠습니까?</ZModal>

<div class="rounded border py-4">
  <div class="px-4">
    <div>
      {#if post}
        <div class="float-left mr-2">
          <span class="rounded border p-1 text-xs text-gray-600 dark:text-gray-400">
            {post.cat}
          </span>
        </div>
      {/if}
      <h3 class="text-xl">{post ? post.title : ''}</h3>
    </div>

    {#if post}
      <div class="py-3">
        <AvatarUser user={{ id: post.user_id, name: post.user_name }} />
        <div class="text-sm">
          <span class="mr-4">{formatDate(post.created_at)}</span>
          <span>조회 {post.hit}</span>
        </div>
      </div>
    {/if}

    <hr />

    {#if isLoading}
      <div class="min-h-36 py-10 text-center text-gray-500">
        <ZSpinner />
      </div>
    {:else if post}
      <div class="min-h-36 py-4">
        <ViewerHTML body={post.body} />
      </div>
      <hr />
      <ViewerReplies {postId} />
    {:else}
      <div class="min-h-36 py-10 text-center text-gray-500">게시글을 불러올 수 없습니다.</div>
    {/if}
  </div>
</div>

<div class="flex gap-3 py-4">
  <ZButton as="a" href={resolve('/forum/new')} disabled={!$canWrite}>글쓰기</ZButton>

  {#if post}
    {#if $canEdit(post.user_id)}
      <ZButton onclick={edit}>수정</ZButton>
    {/if}
    {#if $canDelete(post.user_id)}
      <ZButton onclick={del}>삭제</ZButton>
    {/if}
  {/if}
</div>
