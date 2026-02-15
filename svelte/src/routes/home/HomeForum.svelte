<script lang="ts">
  import { onMount } from 'svelte'

  import { resolve } from '$app/paths'
  import httpy from '$shared/utils/httpy'

  interface Row {
    id: number
    title: string
    replies_count: number
  }

  let rows: Row[] = []

  const load = async () => {
    const [data, err] = await httpy.get<Row[]>('/api/posts/recent')
    if (err) {
      console.error('recent posts', err)
      return
    }

    rows = data
  }

  onMount(load)
</script>

{#each rows as r (r.id)}
  <div class="p-1">
    <a href={resolve(`/forum/${r.id}`)}>
      {r.title}
      {#if r.replies_count > 0}
        <small>[{r.replies_count}]</small>
      {/if}
    </a>
  </div>
{/each}
