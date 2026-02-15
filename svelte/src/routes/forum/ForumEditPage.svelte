<svelte:options runes={true} />

<script lang="ts">
  import './assets/forum-apex.css'

  import { goto } from '$app/navigation'
  import { resolve } from '$app/paths'
  import { page } from '$app/state'
  import ZModal from '$shared/ui/ZModal.svelte'
  import httpy from '$shared/utils/httpy'

  import ForumPostForm from './components/ForumPostForm.svelte'
  import type { ForumPostFormValue } from './components/types'
  import { useErrors } from './errors'

  const errors = useErrors()

  let showCancelModal = $state(false)
  let submitting = $state(false)
  let loading = $state(false)

  let id = $derived(Number(page.params.id || 0))
  let isEdit = $derived(id > 0)

  let form = $state<ForumPostFormValue>({
    cat: '질문',
    title: '',
    body: '',
  })

  let titleError = $derived(errors.has('title') ? errors.get('title').join('') : null)
  let bodyError = $derived(errors.has('body') ? errors.get('body').join('') : null)

  async function fetchData() {
    if (!isEdit) return

    loading = true
    try {
      const [data, err] = await httpy.get<{ cat: string; title: string; body: string }>(`/api/posts/${id}`)
      if (err) {
        console.error('Failed to fetch post:', err)
        return
      }

      form = {
        cat: data.cat ?? '질문',
        title: data.title ?? '',
        body: data.body ?? '',
      }
    } finally {
      loading = false
    }
  }

  $effect(() => {
    if (isEdit) {
      void fetchData()
    }
  })

  function validate() {
    errors.clearAll()

    if (!form.title.trim()) errors.add('title', '제목을 입력해 주세요')
    if (!form.body.trim()) errors.add('body', '내용을 입력해 주세요')

    return !errors.isError()
  }

  function gotoAfterSubmit(newId?: number) {
    goto(resolve(`/forum/${newId ?? id}`))
  }

  async function submit() {
    if (!validate()) return

    submitting = true
    try {
      if (isEdit) {
        const [, err] = await httpy.put(`/api/posts/${id}`, form)
        if (err) {
          console.error('PUT post', err)
          return
        }
        gotoAfterSubmit()
        return
      }

      const [data, err] = await httpy.post<{ id: number }>('/api/posts', form)
      if (err) {
        console.error('POST post', err)
        return
      }
      gotoAfterSubmit(data?.id)
    } finally {
      submitting = false
    }
  }

  function cancel() {
    showCancelModal = true
  }

  function cancelOk() {
    showCancelModal = false
    goto(resolve(isEdit ? `/forum/${id}` : '/forum'))
  }
</script>

<ZModal show={showCancelModal} on:ok={cancelOk} on:cancel={() => (showCancelModal = false)}>
  {isEdit ? '글 수정하기를 취소하시겠습니까?' : '새 글 쓰기를 취소하시겠습니까?'}
</ZModal>

<div class="p-5">
  <div class="container mx-auto max-w-[1140px] px-4">
    <h2 class="my-5 text-2xl font-bold">
      {isEdit ? '포럼 글 수정하기' : '포럼 새 글 쓰기'}
    </h2>

    {#if loading}
      <div class="py-10 text-center text-gray-500">불러오는 중...</div>
    {:else}
      <ForumPostForm
        bind:modelValue={form}
        submitText="저장"
        {submitting}
        {titleError}
        {bodyError}
        onSubmit={submit}
        onCancel={cancel}
        onClearTitleError={() => errors.clear('title')}
        onClearBodyError={() => errors.clear('body')}
      />
    {/if}
  </div>
</div>
