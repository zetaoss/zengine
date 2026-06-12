<script lang="ts">
  import { mdiFileCogOutline } from '@mdi/js'

  import CButton from '$shared/ui/CButton.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'

  export let titles: string[] = []
  export let onSelect: (title: string) => void
  export let onToggleAiEdit: () => void
  export let autoApplyOnSelect = false
  let selected = ''
  let hasUserInteracted = false

  $: {
    if (titles.length > 0 && !titles.includes(selected)) {
      if (autoApplyOnSelect && !hasUserInteracted) {
        selected = titles[0]
        onSelect(selected)
      } else {
        selected = ''
      }
    }
  }

  function handleChange() {
    hasUserInteracted = true
    onSelect(selected)
  }

  function toBoilerplateLabel(title: string) {
    return title.replace(/^틀:새문서틀\s*/, '')
  }
</script>

<div class="mb-2 rounded border border-a-gray-300 px-3 py-2">
  <div class="grid grid-cols-2 gap-3">
    <div class="flex items-center gap-2">
      <span class="text-sm font-medium text-a-slate-700">템플릿:</span>
      <select
        class="cursor-pointer rounded border px-2 py-0.5 text-sm scheme-light-dark hover:bg-a-slate-100"
        bind:value={selected}
        onchange={handleChange}
      >
        <option value="">빈값</option>
        {#each titles as title (title)}
          <option value={title}>{toBoilerplateLabel(title)}</option>
        {/each}
      </select>
      <CButton
        href="/tool/article-tpl"
        target="_blank"
        rel="noopener noreferrer"
        size="small"
        variant="outline"
        class="inline-flex items-center"
        aria-label="템플릿 관리 바로가기"
        title="템플릿 관리"
      >
        <ZIcon path={mdiFileCogOutline} class="h-4 w-4" />
      </CButton>
    </div>

    <div class="flex items-center justify-end">
      <CButton id="zeta-ai-edit-button" size="small" variant="outline" aria-label="AI 편집" onclick={onToggleAiEdit}>AI 편집</CButton>
    </div>
  </div>
</div>
