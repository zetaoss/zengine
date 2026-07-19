<script lang="ts">
  import { onMount } from 'svelte'

  import { subscribeWikiEditorContent } from '../wikiEditor'
  import { compactDiff, createDiff, createSplitRows } from './diff'
  import { selectedAiEditResult } from './selectedAiEditResult'

  const contextLineCount = 3

  let editorContent = $state('')
  let selectedResult = $state<{ content: string; taskId: number } | null>(null)

  let diffLines = $derived(selectedResult ? createDiff(editorContent, selectedResult.content) : [])
  let splitRows = $derived(createSplitRows(diffLines))
  let visibleRows = $derived(compactDiff(splitRows, contextLineCount))
  let hasChanges = $derived(diffLines.some((line) => line.kind !== 'context'))

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
    <div class="flex items-center justify-between border-b border-a-slate-200 bg-a-slate-50 px-3 py-2 text-sm">
      <strong>비교</strong>
      <span class="text-xs text-a-slate-500">#{selectedResult.taskId} · 위키편집기 내용 → AI 편집본</span>
    </div>

    {#if hasChanges}
      <div class="max-h-80 overflow-auto font-mono text-xs leading-5">
        <div class="min-w-[900px]">
          <div class="grid grid-cols-2 border-b border-a-slate-300 bg-a-slate-100 font-sans font-semibold text-a-slate-600">
            <div class="border-r border-a-slate-300 px-3 py-1.5">위키편집기 내용</div>
            <div class="px-3 py-1.5">AI 편집본</div>
          </div>
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
    {:else}
      <div class="px-3 py-4 text-center text-sm text-a-slate-500">위키편집기 내용과 AI 편집본이 같습니다.</div>
    {/if}
  </section>
{/if}
