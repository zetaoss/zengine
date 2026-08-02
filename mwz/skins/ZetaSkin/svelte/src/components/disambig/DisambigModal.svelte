<script lang="ts">
  import { mdiChevronDown, mdiChevronUp, mdiClose, mdiDeleteOutline, mdiDotsVertical, mdiInformationOutline, mdiPlus } from '@mdi/js'
  import { onMount, tick } from 'svelte'

  import mwapi from '$lib/utils/mwapi'
  import CButton from '$shared/ui/CButton.svelte'
  import CMenu from '$shared/ui/CMenu.svelte'
  import CMenuItem from '$shared/ui/CMenuItem.svelte'
  import { showConfirm } from '$shared/ui/confirm/confirm'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import ZModal from '$shared/ui/ZModal.svelte'
  import { showToast } from '$shared/ui/toast/toast'
  import { getWikiViewHref } from '$shared/utils/wikiLink'

  type Row = {
    id: number
    title: string
    displayName: string
    description: string
  }

  type OtherMeaning = {
    title: string
    displayName: string
  }

  interface TokenResponse {
    query?: {
      tokens?: {
        csrftoken?: string
      }
    }
  }

  interface EditResponse {
    edit?: {
      result?: string
    }
  }

  interface AllPagesResponse {
    continue?: {
      apcontinue?: string
    }
    query?: {
      allpages?: Array<{
        title: string
      }>
    }
  }

  interface ContentResponse {
    query?: {
      pages?: Array<{
        missing?: boolean
        redirect?: boolean
        revisions?: Array<{
          slots?: {
            main?: {
              content?: string
            }
          }
          timestamp?: string
        }>
      }>
    }
  }

  export let show = false
  export let baseTitle: string
  export let sourceTitle: string
  export let targetTitle: string
  export let targetExists: boolean
  export let onClose: () => void

  let rows: Row[] = []
  let nextId = 1
  let saving = false
  let deleting = false
  let loading = true
  let autoAdding = false
  let loadFailed = false
  let selectedRowId: number | null = null
  let baseTimestamp: string | undefined
  let tableContainer: HTMLDivElement | null = null

  const newRow = (title = '', displayName = ''): Row => ({
    id: nextId++,
    title,
    displayName,
    description: '',
  })

  const reset = () => {
    nextId = 1
    const firstRow = newRow(sourceTitle)
    rows = [firstRow, newRow()]
    selectedRowId = firstRow.id
    saving = false
    deleting = false
    loading = true
    autoAdding = false
    loadFailed = false
    baseTimestamp = undefined
  }

  const addRow = async () => {
    const row = newRow()
    const selectedIndex = rows.findIndex((item) => item.id === selectedRowId)
    const insertIndex = selectedIndex >= 0 ? selectedIndex : rows.length
    rows = [...rows.slice(0, insertIndex), row, ...rows.slice(insertIndex)]
    selectedRowId = row.id
    await tick()
    tableContainer?.querySelector<HTMLInputElement>(`input[data-row-id="${row.id}"]`)?.focus()
  }

  const removeRow = (id: number) => {
    const removedIndex = rows.findIndex((row) => row.id === id)
    rows = rows.filter((row) => row.id !== id)
    if (selectedRowId === id) {
      selectedRowId = rows[Math.min(removedIndex, rows.length - 1)]?.id ?? null
    }
  }

  const trimRowField = (row: Row, field: 'title' | 'displayName' | 'description') => {
    row[field] = row[field].trim()
    rows = [...rows]
  }

  const moveRow = (index: number, offset: -1 | 1) => {
    const targetIndex = index + offset
    if (targetIndex < 0 || targetIndex >= rows.length) return
    if (!rows[index].title.trim() || !rows[targetIndex].title.trim()) return

    const reordered = [...rows]
    const current = reordered[index]
    reordered[index] = reordered[targetIndex]
    reordered[targetIndex] = current
    rows = reordered
  }

  const moveSelectedRow = (offset: -1 | 1) => {
    const index = rows.findIndex((row) => row.id === selectedRowId)
    if (index >= 0) moveRow(index, offset)
  }

  const removeSelectedRow = () => {
    if (selectedRowId !== null) removeRow(selectedRowId)
  }

  const normalizedRows = (): Row[] =>
    rows
      .map((row) => ({
        ...row,
        title: row.title.trim(),
        displayName: row.displayName.trim(),
        description: row.description.trim().replace(/\s+/g, ' '),
      }))
      .filter((row) => row.title !== '')

  const isValidTitle = (rawTitle: string): boolean => {
    const title = rawTitle.trim()
    if (title === baseTitle) return true

    const suffix = title.slice(baseTitle.length)
    if (!/^ ?\([^()]+\)$/.test(suffix)) return false

    return suffix.slice(suffix.indexOf('(') + 1, -1).trim().length > 0
  }

  const findExistingTitles = async (): Promise<string[] | null> => {
    const found: string[] = []
    let apcontinue: string | undefined

    do {
      const [data, err] = await mwapi.get<AllPagesResponse>({
        action: 'query',
        list: 'allpages',
        apnamespace: 0,
        apprefix: baseTitle,
        apfilterredir: 'nonredirects',
        aplimit: 'max',
        apcontinue,
      })

      if (err) {
        showToast(`관련 문서를 불러오지 못했습니다: ${err.message}`)
        return null
      }

      for (const page of data?.query?.allpages ?? []) {
        if (isValidTitle(page.title) && !found.includes(page.title)) found.push(page.title)
      }
      apcontinue = data?.continue?.apcontinue
    } while (apcontinue)

    return found
  }

  const findOtherMeanings = async (): Promise<OtherMeaning[]> => {
    const [data, err] = await mwapi.get<ContentResponse>({
      action: 'query',
      prop: 'revisions',
      titles: sourceTitle,
      rvprop: 'content',
      rvslots: 'main',
    })
    if (err) {
      showToast(`${sourceTitle} 문서의 다른 뜻 정보를 불러오지 못했습니다: ${err.message}`)
      return []
    }

    const wikitext = data?.query?.pages?.[0]?.revisions?.[0]?.slots?.main?.content ?? ''
    const withoutComments = wikitext.replace(/<!--[\s\S]*?-->/g, '')
    const meanings: OtherMeaning[] = []
    const templatePattern = /\{\{\s*다른[\s_]*뜻\s*\|([^{}]*?)\}\}/giu

    for (const match of withoutComments.matchAll(templatePattern)) {
      const params = match[1].split('|').map((value) => value.trim())
      const first = params[0] ?? ''
      const second = params[1] ?? ''
      const title = second || first
      const displayName = second ? first : ''
      if (!title || !isValidTitle(title)) continue
      if (displayName && !isValidTitle(displayName)) continue
      if (!meanings.some((meaning) => meaning.title === title && meaning.displayName === displayName)) {
        meanings.push({ title, displayName })
      }
    }

    return meanings
  }

  const mergeFoundTitles = (found: string[], meanings: OtherMeaning[]): number => {
    const populated = rows.filter((row) => row.title.trim() !== '')
    const included = populated.map((row) => row.title.trim())
    let added = 0

    for (const meaning of meanings) {
      const existingIndex = included.indexOf(meaning.title)
      if (existingIndex >= 0) {
        const existing = populated[existingIndex]
        if (meaning.displayName && !existing.displayName.trim()) existing.displayName = meaning.displayName
        continue
      }

      populated.push(newRow(meaning.title, meaning.displayName))
      included.push(meaning.title)
      added++
    }

    for (const title of found) {
      if (!included.includes(title)) {
        populated.push(newRow(title))
        included.push(title)
        added++
      }
    }
    if (!included.includes(sourceTitle)) {
      populated.push(newRow(sourceTitle))
      added++
    }

    rows = [...populated, newRow()]
    return added
  }

  const autoAdd = async (notify = true): Promise<boolean> => {
    if (autoAdding) return false

    autoAdding = true
    const [found, meanings] = await Promise.all([findExistingTitles(), findOtherMeanings()])
    autoAdding = false
    if (!found) return false

    const added = mergeFoundTitles(found, meanings)
    if (notify) {
      showToast(added > 0 ? `${added}개 문서를 자동 추가했습니다.` : '자동 추가할 문서가 없습니다.')
    }
    return true
  }

  const parseWikitext = (wikitext: string): Row[] | null => {
    const parsed: Row[] = []

    for (const line of wikitext.split(/\r?\n/)) {
      if (line.trim() === '') continue
      const match = line.match(/^\*\s*\[\[([^\]|]+)(?:\|([^\]]*))?\]\](?:\s*-\s*(.*))?\s*$/)
      if (!match) return null
      parsed.push({
        id: nextId++,
        title: match[1].trim(),
        displayName: (match[2] ?? '').trim(),
        description: (match[3] ?? '').trim(),
      })
    }

    return parsed
  }

  const loadTargetRows = async (): Promise<Row[] | null> => {
    if (!targetExists) return []

    const [data, err] = await mwapi.get<ContentResponse>({
      action: 'query',
      prop: 'revisions',
      titles: targetTitle,
      rvprop: 'content|timestamp',
      rvslots: 'main',
    })
    if (err) {
      showToast(`${targetTitle} 문서를 불러오지 못했습니다: ${err.message}`)
      return null
    }

    const revision = data?.query?.pages?.[0]?.revisions?.[0]
    const parsed = parseWikitext(revision?.slots?.main?.content ?? '')
    if (!parsed) {
      showToast(`${targetTitle} 문서에 표 형식으로 편집할 수 없는 내용이 있습니다.`)
      return null
    }

    baseTimestamp = revision?.timestamp
    return parsed
  }

  const initializeRows = async () => {
    if (!targetExists) {
      const loaded = await autoAdd(false)
      if (!loaded) loadFailed = true
      loading = false
      return
    }

    const targetRows = await loadTargetRows()
    if (!targetRows) {
      loadFailed = true
      loading = false
      return
    }

    const loadedRows = [...targetRows]
    if (!loadedRows.some((row) => row.title.trim() === sourceTitle) && isValidTitle(sourceTitle)) {
      loadedRows.push(newRow(sourceTitle))
    }
    rows = [...loadedRows, newRow()]
    selectedRowId = rows[0]?.id ?? null
    loading = false
  }

  const toWikitext = (items: Row[]): string =>
    items
      .map((row) => {
        const link = row.displayName ? `[[${row.title}|${row.displayName}]]` : `[[${row.title}]]`
        return row.description ? `* ${link} - ${row.description}` : `* ${link}`
      })
      .join('\n')

  const save = async () => {
    if (saving || deleting) return

    const items = normalizedRows()
    if (
      items.length === 0 ||
      items.some((row) => !isValidTitle(row.title)) ||
      items.some((row) => row.displayName && !isValidTitle(row.displayName))
    )
      return

    saving = true
    const [tokenData, tokenErr] = await mwapi.get<TokenResponse>({
      action: 'query',
      meta: 'tokens',
      type: 'csrf',
    })
    const token = tokenData?.query?.tokens?.csrftoken

    if (tokenErr || !token) {
      saving = false
      showToast(`저장 토큰을 가져오지 못했습니다: ${tokenErr?.message ?? 'unknown error'}`)
      return
    }

    const [, editErr] = await mwapi.post<EditResponse>({
      action: 'edit',
      title: targetTitle,
      text: toWikitext(items),
      token,
      createonly: targetExists ? undefined : 1,
      basetimestamp: targetExists ? baseTimestamp : undefined,
      summary: targetExists ? '동음이의 관계 편집' : '동음이의 관계 등록',
    })

    if (editErr) {
      saving = false
      showToast(`동음이의 ${targetExists ? '편집' : '등록'}에 실패했습니다: ${editErr.message}`)
      return
    }

    showToast(`동음이의 ${targetExists ? '편집' : '등록'}이 완료되었습니다.`)
    onClose()
    window.location.reload()
  }

  const deleteTarget = async () => {
    if (!targetExists || saving || deleting) return

    const confirmed = await showConfirm(`'${baseTitle}' 동음이의 문서를 삭제하고 연결을 모두 해제하시겠습니까?`, {
      okText: '삭제',
      okVariant: 'destructive',
    })
    if (!confirmed) return

    deleting = true
    const [tokenData, tokenErr] = await mwapi.get<TokenResponse>({
      action: 'query',
      meta: 'tokens',
      type: 'csrf',
    })
    const token = tokenData?.query?.tokens?.csrftoken

    if (tokenErr || !token) {
      deleting = false
      showToast(`삭제 토큰을 가져오지 못했습니다: ${tokenErr?.message ?? 'unknown error'}`)
      return
    }

    const [, deleteErr] = await mwapi.post<Record<string, never>>({
      action: 'delete',
      title: targetTitle,
      token,
      reason: '동음이의 관계 삭제',
    })

    if (deleteErr) {
      deleting = false
      showToast(`${targetTitle} 문서를 삭제하지 못했습니다: ${deleteErr.message}`)
      return
    }

    showToast(`${targetTitle} 문서와 동음이의 관계를 삭제했습니다.`)
    onClose()
    window.location.reload()
  }

  $: nonEmptyRows = rows.filter((row) => row.title.trim() !== '')
  $: hasInvalidTitle = nonEmptyRows.some((row) => !isValidTitle(row.title))
  $: hasInvalidDisplayName = nonEmptyRows.some((row) => row.displayName.trim() !== '' && !isValidTitle(row.displayName))
  $: canSave = !loadFailed && nonEmptyRows.length > 0 && !hasInvalidTitle && !hasInvalidDisplayName
  $: selectedRowIndex = rows.findIndex((row) => row.id === selectedRowId)
  $: canMoveSelectedUp = selectedRowIndex > 0 && Boolean(rows[selectedRowIndex]?.title.trim() && rows[selectedRowIndex - 1]?.title.trim())
  $: canMoveSelectedDown =
    selectedRowIndex >= 0 &&
    selectedRowIndex < rows.length - 1 &&
    Boolean(rows[selectedRowIndex]?.title.trim() && rows[selectedRowIndex + 1]?.title.trim())

  onMount(() => {
    reset()
    void initializeRows()
  })
</script>

<ZModal
  {show}
  title={targetTitle}
  okText={loading ? '불러오는 중...' : deleting ? '삭제 중...' : saving ? '저장 중...' : '저장'}
  okVariant="default"
  okDisabled={loading || autoAdding || saving || deleting || !canSave}
  backdropClosable={!saving && !deleting}
  closable={!saving && !deleting}
  panelClass="max-w-[calc(100vw-1rem)] md:max-w-5xl"
  sectionClass="min-h-0 flex-1 overflow-x-hidden overflow-y-auto p-3 md:p-5"
  onOk={save}
  onCancel={() => {
    if (!saving && !deleting) onClose()
  }}
>
  <svelte:fragment slot="title">
    <span class="inline-flex items-baseline gap-1">
      {#if targetExists}
        <a class="text-a-sky-600 hover:underline" href={getWikiViewHref(targetTitle)} target="_blank" rel="noopener noreferrer"
          >{baseTitle}</a
        >
      {:else}
        <span>{baseTitle}</span>
      {/if}
      <span>동음이의 {targetExists ? '편집' : '생성'}</span>
      <a
        class="inline-flex self-center rounded p-0.5 text-muted-foreground hover:bg-muted hover:text-foreground"
        href={getWikiViewHref('제타위키:동음이의어_문서')}
        target="_blank"
        rel="noopener noreferrer"
        title="동음이의어 문서 안내"
        aria-label="동음이의어 문서 안내"
      >
        <ZIcon path={mdiInformationOutline} />
      </a>
    </span>
  </svelte:fragment>

  <div class="min-w-0">
    {#if loading}
      <p class="mb-3 text-sm text-muted-foreground">
        {targetExists ? `${targetTitle} 문서를 불러오고 있습니다...` : '관련 문서 제목을 탐색하고 있습니다...'}
      </p>
    {/if}

    <div bind:this={tableContainer} class="mt-0.5 w-full max-w-full overflow-x-auto">
      <table class="w-full table-auto border-collapse text-sm">
        <thead class="bg-muted/40 text-left font-semibold text-muted-foreground">
          <tr>
            <th class="border px-1 py-1 text-center">#</th>
            <th class="border px-2 py-1">제목</th>
            <th class="border px-2 py-1">표시명</th>
            <th class="border px-2 py-1">설명</th>
          </tr>
        </thead>
        <tbody>
          {#each rows as row, index (row.id)}
            <tr class={selectedRowId === row.id ? 'bg-a-sky-100' : ''}>
              <td class="border p-0 text-muted-foreground">
                <button
                  type="button"
                  class={`flex h-8 w-full items-center justify-center font-medium ${selectedRowId === row.id ? '' : 'hover:bg-muted/60'}`}
                  aria-label={`${index + 1}행 선택`}
                  aria-pressed={selectedRowId === row.id}
                  onclick={() => (selectedRowId = row.id)}>{index + 1}</button
                >
              </td>
              <td class="border p-0">
                <input
                  type="text"
                  data-row-id={row.id}
                  size="1"
                  class={`block h-8 min-w-0 w-full rounded-none border-0 bg-transparent px-2 text-sm outline-none placeholder:text-muted-foreground/50 focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring ${row.title.trim() && !isValidTitle(row.title) ? 'ring-1 ring-inset ring-destructive' : ''}`}
                  placeholder={`${baseTitle} (구분)`}
                  bind:value={row.title}
                  onfocus={() => (selectedRowId = row.id)}
                  onblur={() => trimRowField(row, 'title')}
                />
              </td>
              <td class="border p-0">
                <input
                  type="text"
                  size="1"
                  class={`block h-8 min-w-0 w-full rounded-none border-0 bg-transparent px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring ${row.displayName.trim() && !isValidTitle(row.displayName) ? 'ring-1 ring-inset ring-destructive' : ''}`}
                  bind:value={row.displayName}
                  onfocus={() => (selectedRowId = row.id)}
                  onblur={() => trimRowField(row, 'displayName')}
                />
              </td>
              <td class="border p-0">
                <input
                  type="text"
                  size="1"
                  class="block h-8 min-w-0 w-full rounded-none border-0 bg-transparent px-2 text-sm outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
                  bind:value={row.description}
                  onfocus={() => (selectedRowId = row.id)}
                  onblur={() => trimRowField(row, 'description')}
                />
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    {#if hasInvalidTitle}
      <p class="mt-2 text-xs text-destructive">{baseTitle}, {baseTitle}(...), {baseTitle} (...) 형식만 등록할 수 있습니다.</p>
    {/if}
    {#if hasInvalidDisplayName}
      <p class="mt-2 text-xs text-destructive">표시명도 {baseTitle}, {baseTitle}(...), {baseTitle} (...) 형식이어야 합니다.</p>
    {/if}

    <div class="mt-2 flex flex-wrap items-center justify-between gap-2">
      <div class="flex items-center gap-1">
        <CButton
          variant="outline"
          size="sm"
          class="w-14!"
          title="선택한 행을 위로 이동"
          disabled={saving || !canMoveSelectedUp}
          onclick={() => moveSelectedRow(-1)}
        >
          <ZIcon path={mdiChevronUp} />
        </CButton>
        <CButton
          variant="outline"
          size="sm"
          class="w-14!"
          title="선택한 행을 아래로 이동"
          disabled={saving || !canMoveSelectedDown}
          onclick={() => moveSelectedRow(1)}
        >
          <ZIcon path={mdiChevronDown} />
        </CButton>
        <CButton
          variant="outline"
          size="sm"
          class="w-14!"
          title="선택한 행 위치에 추가"
          disabled={loading || autoAdding || saving || selectedRowIndex < 0}
          onclick={addRow}
        >
          <ZIcon path={mdiPlus} />
        </CButton>
        <CButton
          variant="outline"
          size="sm"
          class="w-14!"
          title="선택한 행 삭제"
          disabled={saving || selectedRowIndex < 0}
          onclick={removeSelectedRow}
        >
          <ZIcon path={mdiClose} />
        </CButton>
      </div>
      <div class="flex items-center gap-1">
        <CButton variant="outline" size="sm" disabled={loading || autoAdding || saving || deleting} onclick={() => void autoAdd()}>
          {autoAdding ? '자동 추가 중...' : '자동 추가'}
        </CButton>
        {#if targetExists}
          <CMenu side="top" disabled={loading || saving || deleting}>
            {#snippet trigger({ toggle })}
              <CButton
                variant="outline"
                size="icon-sm"
                title="동음이의 문서 메뉴"
                disabled={loading || saving || deleting}
                onclick={toggle}
              >
                <ZIcon path={mdiDotsVertical} />
              </CButton>
            {/snippet}
            {#snippet menu({ close })}
              <CMenuItem
                onclick={() => {
                  close()
                  void deleteTarget()
                }}
              >
                <span class="inline-flex items-center gap-2 text-destructive">
                  <ZIcon path={mdiDeleteOutline} />
                  동음이의 문서 삭제
                </span>
              </CMenuItem>
            {/snippet}
          </CMenu>
        {/if}
      </div>
    </div>
  </div>
</ZModal>
