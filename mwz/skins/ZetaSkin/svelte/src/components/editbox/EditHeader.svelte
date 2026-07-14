<script lang="ts">
  import { onMount } from 'svelte'

  import getCurrentTitle from '$lib/utils/getCurrentTitle'
  import mwapi from '$lib/utils/mwapi'
  import getRLCONF from '$lib/utils/rlconf'
  import CButton from '$shared/ui/CButton.svelte'
  import ZNativeSelect from '$shared/ui/ZNativeSelect.svelte'
  import ZToggle from '$shared/ui/ZToggle.svelte'
  import { getAge } from '$shared/utils/time'

  import { getWikiEditorContent, replaceWikiEditorContent } from './wikiEditor'

  interface Revision {
    revid: number
    parentid: number
    user: string
    timestamp: string
    comment?: string
  }

  interface LoadSelectItem {
    group: string
    label: string
    value: string
  }

  export let titles: string[] = []
  export let onSelect: (title: string) => void
  export let onToggleAiEdit: (visible: boolean) => void
  export let autoApplyOnSelect = false

  let selectedValue = '' // e.g. "template:틀:..." or "revision:12345"
  let hasUserInteracted = false
  let originalContent = ''
  let revisions: Revision[] = []
  let aiEditVisible = false
  let loadSelectItems: LoadSelectItem[]
  const isSysop = (getRLCONF()?.wgUserGroups || []).includes('sysop')

  $: loadSelectItems = [
    ...titles.map((title) => ({ group: 'Templates', label: toBoilerplateLabel(title), value: `template:${title}` })),
    ...revisions.map((revision) => ({ group: 'History', label: formatRevisionLabel(revision), value: `revision:${revision.revid}` })),
  ]

  $: {
    if (titles.length > 0) {
      const isTemplateSelected = selectedValue.startsWith('template:')
      const selectedTitle = isTemplateSelected ? selectedValue.substring('template:'.length) : ''

      if (!titles.includes(selectedTitle)) {
        if (autoApplyOnSelect && !hasUserInteracted) {
          const newSelectedTitle = titles[0]
          selectedValue = `template:${newSelectedTitle}`
          onSelect(newSelectedTitle)
        } else if (!selectedValue.startsWith('revision:')) {
          selectedValue = ''
        }
      }
    }
  }

  async function fetchRevisions() {
    const title = getCurrentTitle()
    if (!title) return

    const [data, err] = await mwapi.get<{
      query: {
        pages: Array<{
          revisions?: Revision[]
        }>
      }
    }>({
      action: 'query',
      prop: 'revisions',
      titles: title,
      rvlimit: 10,
      rvprop: 'ids|timestamp|user|comment',
    })

    if (err) {
      console.error('Failed to fetch revisions:', err)
      return
    }

    const pages = data?.query?.pages
    if (pages && Array.isArray(pages) && pages.length > 0) {
      revisions = pages[0].revisions || []
    }
  }

  onMount(() => {
    originalContent = getWikiEditorContent()
    void fetchRevisions()
  })

  async function handleSelectChange(nextValue: string) {
    selectedValue = nextValue
    hasUserInteracted = true
    if (!selectedValue) {
      onSelect('')
      return
    }

    const [type, ...rest] = selectedValue.split(':')
    const value = rest.join(':')

    if (type === 'template') {
      onSelect(value)
    } else if (type === 'revision') {
      onSelect('') // Selecting a revision should clear any selected template.
      const revId = Number(value)
      if (!revId) return

      const [data, err] = await mwapi.get<{
        query: {
          pages: Array<{
            revisions?: Array<{
              slots?: {
                main?: {
                  content?: string
                }
              }
              content?: string
            }>
          }>
        }
      }>({
        action: 'query',
        prop: 'revisions',
        revids: String(revId),
        rvprop: 'content',
        rvslots: 'main',
      })

      if (err) {
        console.error('Failed to fetch revision content:', err)
        return
      }

      const pages = data?.query?.pages
      if (pages && Array.isArray(pages) && pages.length > 0) {
        const revs = pages[0].revisions
        if (revs && Array.isArray(revs) && revs.length > 0) {
          const content = revs[0].slots?.main?.content ?? revs[0].content ?? ''
          replaceWikiEditorContent(content)
        }
      }
    }
  }

  function handleReset() {
    replaceWikiEditorContent(originalContent)
    selectedValue = ''
    onSelect('')
  }

  function handleAiEditToggle(event: { checked: boolean }) {
    aiEditVisible = event.checked
    onToggleAiEdit(aiEditVisible)
  }

  function toBoilerplateLabel(title: string) {
    return title.replace(/^틀:새문서틀\s*/, '')
  }

  function formatRevisionLabel(rev: Revision) {
    const age = getAge(rev.timestamp)
    let commentStr = rev.comment || ''
    if (commentStr.length > 25) {
      commentStr = commentStr.slice(0, 24) + '...'
    }
    const commentSuffix = commentStr ? ` - ${commentStr}` : ''
    return `r${rev.revid} (${age}) ${rev.user}${commentSuffix}`
  }
</script>

<div class="wikiEditor-ui-toolbar px-2 border-b border-[#ccc]">
  <div class:grid-cols-2={isSysop} class:grid-cols-1={!isSysop} class="grid gap-3">
      <div class="flex items-center gap-2">
      <CButton size="small" variant="outline" aria-label="원복" onclick={handleReset}>원복</CButton>
      <ZNativeSelect
        bind:value={selectedValue}
        items={loadSelectItems}
        placeholder="-- 불러오기 --"
        class="w-full max-w-[150px] text-sm md:max-w-[320px]"
        onchange={(value) => void handleSelectChange(value)}
      />
    </div>

    {#if isSysop}
      <div id="zeta-ai-edit-button" class="flex items-center justify-end gap-2">
        <span class="text-sm text-a-slate-700">AI</span>
        <ZToggle checked={aiEditVisible} label="AI" onchange={handleAiEditToggle} />
      </div>
    {/if}
  </div>
</div>
