<script lang="ts">
  import { onMount } from 'svelte'

  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import ZSkeleton from '$shared/ui/ZSkeleton.svelte'
  import httpy from '$shared/utils/httpy'
  import linkify from '$shared/utils/linkify'

  interface Row {
    id: number
    page_id: number
    page_title: string
    message: string

    user_id: number
    user_name: string
    loading?: boolean
  }

  let rows: Row[] = []

  const load = async () => {
    const [data, err] = await httpy.get<Row[]>('/api/comments/recent')
    if (err) {
      console.error('recent comments', err)
      return
    }

    const safeRows = data ?? []
    rows = safeRows.map((x) => ({
      ...x,
      message: '',
      loading: true,
    }))

    const linkedMessages = await linkify(safeRows.map((x) => x.message || ''))
    rows = safeRows.map((x, i) => ({
      ...x,
      message: linkedMessages[i] ?? '',
      loading: false,
    }))
  }

  onMount(load)
</script>

{#each rows as r (r.id)}
  <div class="py-2">
    <a href={`/wiki/${r.page_title}`} rel="external" data-sveltekit-reload>{r.page_title.replace(/_/g, ' ')}</a>
    <span class="silver ml-3">
      <AvatarUser user={{ id: r.user_id, name: r.user_name }} />
    </span>
    <div class="line-clamp-3 wrap-break-word text-ellipsis">
      {#if r.loading}
        <ZSkeleton width="12rem" height="0.875rem" />
      {:else}
        <!-- eslint-disable-next-line svelte/no-at-html-tags -->
        {@html r.message}
      {/if}
    </div>
  </div>
{/each}
