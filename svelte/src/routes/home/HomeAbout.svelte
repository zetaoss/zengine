<script lang="ts">
  import { onMount } from 'svelte'

  import httpy from '$shared/utils/httpy'

  interface Data {
    query?: {
      statistics?: {
        articles?: number
      }
    }
  }

  let count = ''

  const load = async () => {
    const [data, err] = await httpy.get<Data>('/w/api.php', {
      format: 'json',
      action: 'query',
      meta: 'siteinfo',
      siprop: 'statistics',
    })
    if (err) {
      console.error('articles', err)
      return
    }

    count = data?.query?.statistics?.articles?.toLocaleString() ?? '-'
  }

  onMount(load)
</script>

<div class="p-1">
  <b>'세상의 각주'</b>
  <a class="ml-1" href="/wiki/제타위키" rel="external" data-sveltekit-reload>제타위키</a>에 오신 것을 환영합니다! 누구나 편집할 수 있는
  위키입니다.
  <ul>
    <li>
      글 개수: <a href="/wiki/특수:통계" rel="external" data-sveltekit-reload>{count}</a>
    </li>
    <li>라이선스: CC BY-SA 3.0</li>
    <li>현재 작성 중인 문서는 <a href="/wiki/특수:최근바뀜" rel="external" data-sveltekit-reload>바뀐글</a>을 참조해 주세요.</li>
    <li><a href="/wiki/제타위키_사용법" rel="external" data-sveltekit-reload>제타위키 사용법</a></li>
  </ul>
</div>
