<script lang="ts">
  import { mdiDelete } from '@mdi/js'
  import { onMount } from 'svelte'

  import { useOnelineDelete } from '$lib/composables/useOnelineDelete'
  import useAuthStore from '$lib/stores/auth'
  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import { showToast } from '$shared/ui/toast/toast'
  import ZButton from '$shared/ui/ZButton.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import httpy from '$shared/utils/httpy'
  import linkify from '$shared/utils/linkify'

  interface Row {
    id: number
    user_id: number
    user_name: string
    created: string
    message: string
  }

  let rows: Row[] = []
  let message = ''
  let isSubmitting = false

  const auth = useAuthStore()
  const canWrite = auth.canWrite
  const canDelete = auth.canDelete

  $: trimmedMessage = message.trim()
  $: canSubmit = trimmedMessage.length > 0 && !isSubmitting

  const { del } = useOnelineDelete({
    onSuccess: (row) => {
      rows = rows.filter((item) => item.id !== row.id)
    },
  })

  const load = async () => {
    const [data, err] = await httpy.get<Row[]>('/api/onelines/recent')
    if (err) {
      console.error(err)
      return
    }

    rows = await Promise.all(
      data.map(async (r) => ({
        ...r,
        message: (await linkify([r.message]))[0] ?? '',
      })),
    )
  }

  const submit = async () => {
    if (!$canWrite) {
      showToast('작성 불가')
      return
    }
    if (!canSubmit) return

    isSubmitting = true
    const [data, err] = await httpy.post<Row>('/api/onelines', {
      message: trimmedMessage,
    })
    if (err) {
      console.error('create oneline', err)
      showToast(err.message || '등록 실패')
      isSubmitting = false
      return
    }

    rows = [
      {
        ...data,
        message: (await linkify([data.message]))[0] ?? '',
      },
      ...rows,
    ]
    message = ''
    isSubmitting = false
  }

  const notifyLogin = () => {
    if (!$canWrite) {
      showToast('로그인하면 글을 쓸 수 있어요.')
    }
  }

  onMount(load)
</script>

<form class="mb-2 flex items-center gap-2" on:submit|preventDefault={submit}>
  <div class="relative flex-1">
    <input
      bind:value={message}
      type="text"
      class="w-full rounded border p-1 px-2"
      placeholder="What’s on your mind?"
      disabled={!$canWrite || isSubmitting}
    />
    {#if !$canWrite}
      <button type="button" aria-label="login required" class="absolute inset-0 z-10 cursor-not-allowed" on:click={notifyLogin}></button>
    {/if}
  </div>
  <button type="submit" class="rounded border px-3 py-1 text-sm disabled:opacity-40" disabled={!$canWrite || !canSubmit}> 등록 </button>
</form>

{#each rows as r (r.id)}
  <div class="py-2">
    <AvatarUser user={{ id: r.user_id, name: r.user_name }} />
    <!-- eslint-disable-next-line svelte/no-at-html-tags -->
    <span class="ml-1">{@html r.message}</span>
    <span class="z-muted2 ml-1 text-xs">{r.created.substring(0, 10)}</span>
    {#if $canDelete(r.user_id)}
      <ZButton color="ghost" class="z-muted3 py-1 align-middle leading-none" on:click={() => del(r)}>
        <ZIcon path={mdiDelete} />
      </ZButton>
    {/if}
  </div>
{/each}
