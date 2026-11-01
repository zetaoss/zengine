<script lang="ts">
  import { onMount } from 'svelte'

  import getRLCONF from '$lib/utils/rlconf'
  import { showToast } from '$shared/ui/toast/toast'
  import httpy from '$shared/utils/httpy'

  interface Props {
    pageId?: number
  }

  const { pageId: propPageId }: Props = $props()

  const EMOJIS = ['👍', '❤️', '🎉', '🚀', '👀', '😄'] as const
  type Emoji = (typeof EMOJIS)[number]

  const conf = getRLCONF() ?? {}
  const pageId = $derived(propPageId ?? conf.wgArticleId ?? 0)
  const isLoggedIn = (conf.wgUserId ?? 0) > 0
  const loginUrl = conf.menu?.usermenu?.login?.href ?? '/Special:UserLogin'

  let emojiCounts = $state<Record<string, number>>({})
  let userEmojis = $state<string[]>([])
  let pendingEmoji = $state<string | null>(null)
  let initialLoaded = $state(false)
  let fetchToken = 0

  async function fetchReactions() {
    if (!pageId) return
    const token = ++fetchToken
    const [data, err] = await httpy.get<{
      emojiCount: Record<string, number>
      userEmojis: string[]
    }>(`/api/reactions/page/${pageId}`)

    if (err) {
      console.warn('Failed to load reactions', err)
      initialLoaded = true
      return
    }

    if (token !== fetchToken || !data) return
    emojiCounts = data.emojiCount ?? {}
    userEmojis = data.userEmojis ?? []
    initialLoaded = true
  }

  async function toggleReaction(emoji: Emoji) {
    if (!isLoggedIn) {
      showToast('반응을 남기려면 로그인이 필요합니다.', {
        action: {
          label: '로그인',
          href: loginUrl,
        },
      })
      return
    }

    if (!initialLoaded || pendingEmoji) return

    const currentlyActive = userEmojis.includes(emoji)
    const nextEnable = !currentlyActive

    pendingEmoji = emoji
    const prevCounts = { ...emojiCounts }
    const prevUserEmojis = [...userEmojis]

    // Optimistic UI update
    if (nextEnable) {
      userEmojis = [...userEmojis, emoji]
      emojiCounts = {
        ...emojiCounts,
        [emoji]: (emojiCounts[emoji] || 0) + 1,
      }
    } else {
      userEmojis = userEmojis.filter((e) => e !== emoji)
      const nextCount = (emojiCounts[emoji] || 0) - 1
      if (nextCount > 0) {
        emojiCounts = { ...emojiCounts, [emoji]: nextCount }
      } else {
        const copy = { ...emojiCounts }
        delete copy[emoji]
        emojiCounts = copy
      }
    }

    // Invalidate any in-flight GET so it doesn't overwrite optimistic mutation
    fetchToken++

    const [, err] = await httpy.post('/api/reactions/page', {
      pageid: pageId,
      emoji,
      enable: nextEnable,
    })

    pendingEmoji = null

    if (err) {
      console.error('Failed to toggle reaction', err)
      emojiCounts = prevCounts
      userEmojis = prevUserEmojis
      showToast('반응 저장에 실패했습니다.')
    }
  }

  onMount(() => {
    fetchReactions()
  })
</script>

<div class="flex flex-wrap items-center gap-2 py-2">
  {#each EMOJIS as emoji (emoji)}
    {@const count = emojiCounts[emoji] ?? 0}
    {@const active = userEmojis.includes(emoji)}
    <button
      type="button"
      disabled={!initialLoaded}
      class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full border text-sm transition-colors cursor-pointer select-none
        {!initialLoaded ? 'opacity-60 cursor-not-allowed' : ''}
        {active
        ? 'bg-primary/10 border-primary text-primary font-medium hover:bg-primary/15'
        : 'border-border/60 hover:bg-muted/60 text-foreground/80 hover:border-border'}"
      aria-pressed={active}
      aria-label={`${emoji} 반응 ${count > 0 ? count + '개' : ''}`}
      onclick={() => toggleReaction(emoji)}
    >
      <span class="text-base leading-none">{emoji}</span>
      {#if count > 0}
        <span class="text-xs font-semibold tabular-nums {active ? 'text-primary' : 'text-muted-foreground'}">
          {count}
        </span>
      {/if}
    </button>
  {/each}
</div>
