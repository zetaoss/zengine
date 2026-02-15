<script lang="ts">
  /* eslint-disable svelte/no-navigation-without-resolve */
  import { onMount } from 'svelte'

  import httpy from '$shared/utils/httpy'

  interface Row {
    title: string
  }

  interface Data {
    query?: {
      recentchanges: Row[]
    }
  }

  let rows: Row[] = []

  const load = async () => {
    const [data, err] = await httpy.get<Data>('/w/api.php', {
      format: 'json',
      action: 'query',
      list: 'recentchanges',
      rcprop: 'title',
      rcnamespace: 0,
      rctype: 'new',
      rcshow: '!bot|!anon',
      rclimit: 25,
    })
    if (err) {
      console.error('recentchanges', err)
      return
    }
    rows = data?.query?.recentchanges ?? []
  }

  onMount(load)
</script>

<ul class="py-2 pl-5">
  {#each rows as r (r.title)}
    <li>
      <a href="/wiki/{r.title}" data-sveltekit-reload>{r.title}</a>
    </li>
  {/each}
</ul>
