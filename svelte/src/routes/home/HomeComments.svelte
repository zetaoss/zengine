<script lang="ts">
  import { onMount } from 'svelte'

  import AvatarUser from '$shared/components/avatar/AvatarUser.svelte'
  import httpy from '$shared/utils/httpy'
  import linkify from '$shared/utils/linkify'

  interface Row {
    page_title: string
    message: string
    page_name: string

    user_id: number
    user_name: string
  }

  let rows: Row[] = []

  const load = async () => {
    const [data, err] = await httpy.get<Row[]>('/api/comments/recent')
    if (err) {
      console.error('recent comments', err)
      return
    }

    rows = await Promise.all(
      data.map(async (x) => ({
        ...x,
        message: await linkify(x.message),
      })),
    )
  }

  onMount(load)
</script>

{#each rows as r, i (r.page_title + '-' + r.user_id + '-' + r.message + '-' + i)}
  <div class="py-2">
    <a href={`/wiki/${r.page_title}`} rel="external" data-sveltekit-reload>{r.page_title.replace(/_/g, ' ')}</a>
    <span class="silver ml-3">
      <AvatarUser user={{ id: r.user_id, name: r.user_name }} />
    </span>
    <!-- eslint-disable-next-line svelte/no-at-html-tags -->
    <div class="line-clamp-3 wrap-break-word text-ellipsis">{@html r.message}</div>
  </div>
{/each}
