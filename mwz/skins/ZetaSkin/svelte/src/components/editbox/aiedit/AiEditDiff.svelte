<script lang="ts">
  import { mdiSend } from '@mdi/js'
  import { onMount } from 'svelte'

  import CButton from '$shared/ui/CButton.svelte'
  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  import { replaceWikiEditorContent, subscribeWikiEditorContent } from '../wikiEditor'
  import { compactDiff, createDiff, createSplitRows } from './diff'
  import { selectedAiEditResult, type SelectedAiEditResult } from './selectedAiEditResult'

  const contextLineCount = 3

  let editorContent = $state('')
  let selectedResult = $state<SelectedAiEditResult | null>(null)

  let diffLines = $derived(selectedResult ? createDiff(editorContent, selectedResult.content) : [])
  let splitRows = $derived(createSplitRows(diffLines))
  let visibleRows = $derived(compactDiff(splitRows, contextLineCount))
  let hasChanges = $derived(diffLines.some((line) => line.kind !== 'context'))

  function applyResultToWikiEditor() {
    if (!selectedResult || !replaceWikiEditorContent(selectedResult.content)) {
      showToast('반영할 AI 작성본이 없습니다.')
      return
    }

    if (selectedResult.requestType === 'create') {
      selectedResult.onRequestTypeChange?.('edit')
      showToast('AI 작성본을 위키편집기에 반영했습니다.')
    } else {
      showToast('AI 작성본을 위키편집기에 반영했습니다.')
    }
  }

  onMount(() => {
    const unsubscribe = selectedAiEditResult.subscribe((value) => {
      selectedResult = value
    })
    const unsubscribeEditor = subscribeWikiEditorContent((content) => {
      editorContent = content
    })

    return () => {
      unsubscribe()
      unsubscribeEditor()
    }
  })
</script>

{#if selectedResult}
  <section class="mb-2 overflow-hidden rounded border border-a-slate-300 bg-a-white">
    <div class="h-8 grid grid-cols-2 place-items-center border-b border-a-slate-300 bg-a-slate-50 text-sm">
      <div>위키편집기 내용</div>
      <div>AI 작성본</div>
    </div>
    {#if hasChanges}
      <div class="relative">
        <div class="max-h-80 overflow-auto font-mono text-xs leading-5">
          <div class="min-w-[900px]">
            {#each visibleRows as row, index (`${row?.oldLine?.oldLine ?? ''}-${row?.newLine?.newLine ?? ''}-${index}`)}
              {#if row === null}
                <div class="border-y border-a-slate-200 bg-a-slate-50 px-3 py-1 text-center text-a-slate-400">⋯</div>
              {:else}
                <div class="grid grid-cols-2 border-b border-a-slate-100 last:border-b-0">
                  <div
                    class:bg-a-red-50={row.oldLine?.kind === 'removed'}
                    class:text-a-red-800={row.oldLine?.kind === 'removed'}
                    class="grid grid-cols-[3rem_1.5rem_minmax(0,1fr)] border-r border-a-slate-300"
                  >
                    <span class="select-none border-r border-a-slate-200 px-2 text-right text-a-slate-400">{row.oldLine?.oldLine ?? ''}</span>
                    <span class="select-none text-center">{row.oldLine?.kind === 'removed' ? '−' : ' '}</span>
                    <span class="whitespace-pre-wrap wrap-break-word pr-3">
                      {#if row.oldInline}
                        {#each row.oldInline as segment, segmentIndex (`${segmentIndex}-${segment.changed}`)}
                          <span class:bg-a-red-200={segment.changed}>{segment.text}</span>
                        {/each}
                      {:else}
                        {row.oldLine?.text || ' '}
                      {/if}
                    </span>
                  </div>
                  <div
                    class:bg-a-emerald-50={row.newLine?.kind === 'added'}
                    class:text-a-emerald-800={row.newLine?.kind === 'added'}
                    class="grid grid-cols-[3rem_1.5rem_minmax(0,1fr)]"
                  >
                    <span class="select-none border-r border-a-slate-200 px-2 text-right text-a-slate-400">{row.newLine?.newLine ?? ''}</span>
                    <span class="select-none text-center">{row.newLine?.kind === 'added' ? '+' : ' '}</span>
                    <span class="whitespace-pre-wrap wrap-break-word pr-3">
                      {#if row.newInline}
                        {#each row.newInline as segment, segmentIndex (`${segmentIndex}-${segment.changed}`)}
                          <span class:bg-a-emerald-200={segment.changed}>{segment.text}</span>
                        {/each}
                      {:else}
                        {row.newLine?.text || ' '}
                      {/if}
                    </span>
                  </div>
                </div>
              {/if}
            {/each}
          </div>
        </div>
        <CButton
          type="button"
          variant="subtle"
          size="xs"
          class="absolute right-5 top-2 z-10 shadow-sm"
          disabled={!selectedResult.content}
          onclick={applyResultToWikiEditor}
        >
          <ZIcon path={mdiSend} />
          Apply
        </CButton>
      </div>
    {:else}
      <div class="px-3 py-4 text-center text-sm text-a-slate-500">위키편집기 내용과 AI 작성본이 같습니다.</div>
    {/if}
  </section>
{/if}
