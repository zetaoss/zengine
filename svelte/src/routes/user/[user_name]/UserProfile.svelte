<svelte:options runes={true} />

<script lang="ts">
  import { mdiVectorDifference } from '@mdi/js'
  import { onMount } from 'svelte'
  import { SvelteDate } from 'svelte/reactivity'

  import { resolve } from '$app/paths'
  import { page } from '$app/state'
  import useAuthStore from '$lib/stores/auth'
  import AvatarIcon from '$shared/components/avatar/AvatarIcon.svelte'
  import ZCard from '$shared/ui/ZCard.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZSpinner from '$shared/ui/ZSpinner.svelte'
  import httpy from '$shared/utils/httpy'

  import ContributionMap from './ContributionMap.svelte'

  interface Contribution {
    timestamp: string
    title: string
    revid: number
  }

  interface UserContribsResponse {
    usercontribs?: Contribution[]
  }

  interface ActionQueryResponse<T> {
    query?: T
    error?: { info?: string }
  }

  interface UserData {
    user_id: number
    user_name: string
    user_registration: string
    user_editcount: number
  }

  type StatsMap = Record<string, number>

  const auth = useAuthStore()
  const { isLoggedIn, userInfo } = auth

  let userId = $state(0)
  let editCount = $state(0)
  let minDate = $state(new Date())
  let stats = $state<StatsMap | null>(null)
  let contribs = $state<Contribution[]>([])

  let isLoading = $state(false)
  let loadError = $state<string | null>(null)

  let userName = $derived(page.params.user_name ?? '')
  let encodedUsername = $derived(userName.replace(/ /g, '_'))

  let userPageHref = $derived(`/wiki/User:${encodedUsername}`)
  let contribPageHref = $derived(`/wiki/특수:기여/${encodedUsername}`)
  let isMe = $derived($isLoggedIn && ($userInfo?.id ?? 0) === userId)

  const today = new SvelteDate()
  today.setHours(12, 0, 0, 0)

  function formatYmd(d: Date): string {
    const y = d.getFullYear()
    const m = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${y}-${m}-${day}`
  }

  function agoDate(d: Date): string {
    const diffSec = Math.floor((Date.now() - d.getTime()) / 1000)
    if (diffSec < 60) return `${diffSec}초 전`

    const min = Math.floor(diffSec / 60)
    if (min < 60) return `${min}분 전`

    const hour = Math.floor(min / 60)
    if (hour < 24) return `${hour}시간 전`

    return formatYmd(d)
  }

  function parseRegistrationDate(reg: string): Date {
    if (!reg || reg.length < 8) return new Date()

    const y = Number(reg.slice(0, 4))
    const m = Number(reg.slice(4, 6))
    const d = Number(reg.slice(6, 8))

    if (!Number.isFinite(y) || !Number.isFinite(m) || !Number.isFinite(d)) {
      return new Date()
    }

    return new Date(y, m - 1, d, 12)
  }

  async function fetchOrSetError<T>(resource: string, promise: Promise<[T, null] | [null, unknown]>) {
    const [data, err] = await promise
    if (err) {
      console.error(err)
      loadError = `failed to get ${resource}`
      return null
    }
    return data
  }

  async function load(targetUserName: string) {
    loadError = null
    contribs = []
    stats = null

    if (!$isLoggedIn) await auth.update()

    const userData = await fetchOrSetError<UserData>('UserData', httpy.get(`/api/user/${encodeURIComponent(targetUserName)}`))
    if (!userData) return

    userId = userData.user_id
    editCount = userData.user_editcount
    minDate = parseRegistrationDate(userData.user_registration)

    const contribRes = await fetchOrSetError<ActionQueryResponse<UserContribsResponse>>(
      'UserContribs',
      httpy.get('/w/api.php', {
        action: 'query',
        format: 'json',
        list: 'usercontribs',
        ucuser: targetUserName,
        uclimit: '10',
        ucprop: 'ids|title|timestamp',
      }),
    )
    if (!contribRes) return

    contribs = contribRes.query?.usercontribs ?? []

    const rawStats = await fetchOrSetError<StatsMap>('StatsMap', httpy.get(`/api/user/${userId}/stats`))
    if (!rawStats) return

    stats = rawStats
  }

  let mounted = $state(false)
  let lastLoadedUserName = ''

  onMount(() => {
    mounted = true
    return () => {
      mounted = false
    }
  })

  $effect(() => {
    const currentUserName = userName.trim()
    if (!mounted || currentUserName === '' || currentUserName === lastLoadedUserName) return

    lastLoadedUserName = currentUserName
    isLoading = true
    loadError = null

    void load(currentUserName).finally(() => {
      if (mounted && lastLoadedUserName === currentUserName) {
        isLoading = false
      }
    })
  })
</script>

<div class="mx-auto max-w-4xl py-6">
  {#if isLoading}
    <div class="text-center">
      <ZSpinner />
    </div>
  {:else if loadError}
    <div class="mt-4 text-center text-sm text-red-500">
      {loadError}
    </div>
  {:else}
    <ZCard class="mt-4 p-6">
      <div class="flex flex-row">
        <div class="flex w-1/3 items-center justify-center p-6">
          <AvatarIcon user={{ id: userId, name: userName }} size={96} />
        </div>

        <div class="flex-1 border-l p-6">
          <div class="flex items-baseline gap-2">
            <h1 class="text-xl font-semibold">{userName}</h1>
            {#if isMe}
              <a href={resolve(`/user/${encodedUsername}/edit`)} class="ml-1 text-xs">Edit Profile</a>
            {/if}
          </div>

          <div class="z-muted2 mt-2 space-y-1 text-sm">
            <p>
              편집수:
              <span class="z-text font-semibold">{editCount.toLocaleString()}</span>
            </p>
            <p>
              가입일:
              {formatYmd(minDate)}
            </p>
            <p>
              <a href={userPageHref} rel="external" data-sveltekit-reload>사용자 문서 바로가기</a>
            </p>
          </div>
        </div>
      </div>
    </ZCard>

    <ZCard class="mt-4 p-6">
      <svelte:fragment slot="header">
        <a href={contribPageHref} rel="external" data-sveltekit-reload>최근 편집</a>
      </svelte:fragment>

      <table class="mytable z-muted2 w-full">
        <thead>
          <tr class="border-b">
            <th>일시</th>
            <th>문서</th>
            <th>차이</th>
          </tr>
        </thead>
        <tbody>
          {#each contribs as c (c.revid)}
            <tr class="border-b">
              <td>{agoDate(new Date(c.timestamp))}</td>
              <td>
                <a href={`/wiki/${encodeURIComponent(c.title.replace(/ /g, '_'))}`} rel="external" data-sveltekit-reload>
                  {c.title}
                </a>
              </td>
              <td>
                <a
                  href={`/w/index.php?title=${encodeURIComponent(c.title)}&diff=prev&oldid=${c.revid}`}
                  rel="external"
                  data-sveltekit-reload
                >
                  <ZIcon path={mdiVectorDifference} />
                </a>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </ZCard>

    {#if stats}
      <ContributionMap {stats} {minDate} maxDate={today} />
    {/if}
  {/if}
</div>

<style>
  .mytable td,
  .mytable th {
    padding: 2px 0;
    text-align: left;
  }
</style>
