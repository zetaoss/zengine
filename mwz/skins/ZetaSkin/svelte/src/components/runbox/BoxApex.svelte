<!-- BoxApex.svelte -->
<script lang="ts">
  import { mdiCheck, mdiContentCopy } from '@mdi/js'
  import type { Writable } from 'svelte/store'

  import ZIcon from '$shared/ui/ZIcon.svelte'

  import BoxFront from './BoxFront.svelte'
  import BoxLang from './BoxLang.svelte'
  import BoxNotebook from './BoxNotebook.svelte'
  import BoxZero from './BoxZero.svelte'
  import type { Job } from './types'

  export let job: Writable<Job>
  export let seq: number
  export let contentHtml = ''

  const componentMap = {
    front: BoxFront,
    lang: BoxLang,
    notebook: BoxNotebook,
    zero: BoxZero,
  } as const

  $: jobValue = $job
  $: CurrentComponent = componentMap[jobValue.type as keyof typeof componentMap] ?? BoxZero

  type SimpleBox = { lang: string; text: string }

  $: box = (jobValue.boxes?.[seq] as SimpleBox | undefined) ?? { lang: '', text: '' }
  $: renderedContentHtml = box.lang === 'css' ? enhanceCssColorPreview(contentHtml) : contentHtml

  let copied = false
  let copyTimer: number | null = null

  const onCopy = async () => {
    try {
      await navigator.clipboard.writeText(box.text)
      copied = true
      if (copyTimer) window.clearTimeout(copyTimer)
      copyTimer = window.setTimeout(() => {
        copied = false
        copyTimer = null
      }, 1500)
    } catch (e) {
      console.error(e)
    }
  }

  const CSS_COLOR_TOKEN = /#[0-9a-fA-F]{3,8}\b|rgba?\([^)\n]+\)|hsla?\([^)\n]+\)|\b[a-zA-Z-]+\b/g
  const SKIP_COLOR_KEYWORDS = ['inherit', 'initial', 'unset', 'revert', 'revert-layer', 'transparent', 'currentcolor']

  function isCssColorToken(token: string): boolean {
    if (typeof CSS === 'undefined' || typeof CSS.supports !== 'function') return false
    const normalized = token.trim()
    if (!normalized) return false
    if (SKIP_COLOR_KEYWORDS.includes(normalized.toLowerCase())) return false
    return CSS.supports('color', normalized)
  }

  function enhanceCssColorPreview(sourceHtml: string): string {
    if (typeof document === 'undefined') return sourceHtml
    if (!sourceHtml) return sourceHtml

    const root = document.createElement('div')
    root.innerHTML = sourceHtml

    const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT)
    const textNodes: Text[] = []

    while (walker.nextNode()) {
      textNodes.push(walker.currentNode as Text)
    }

    for (const textNode of textNodes) {
      const text = textNode.data
      const fragments = document.createDocumentFragment()
      let lastIndex = 0
      let hasReplacement = false
      CSS_COLOR_TOKEN.lastIndex = 0
      let match = CSS_COLOR_TOKEN.exec(text)

      while (match) {
        const token = match[0]
        const index = match.index

        if (isCssColorToken(token)) {
          if (index > lastIndex) {
            fragments.appendChild(document.createTextNode(text.slice(lastIndex, index)))
          }

          const swatch = document.createElement('span')
          swatch.className = 'z-color-preview-swatch'
          swatch.style.backgroundColor = token
          swatch.title = token

          fragments.appendChild(swatch)
          fragments.appendChild(document.createTextNode(token))

          lastIndex = index + token.length
          hasReplacement = true
        }

        match = CSS_COLOR_TOKEN.exec(text)
      }

      if (!hasReplacement) continue
      if (lastIndex < text.length) {
        fragments.appendChild(document.createTextNode(text.slice(lastIndex)))
      }

      textNode.parentNode?.replaceChild(fragments, textNode)
    }

    return root.innerHTML
  }
</script>

{#if jobValue}
  <div>
    <div class="mb-1 bg-(--code-bg) p-1 border rounded-lg">
      <svelte:component this={CurrentComponent} {job} {seq}>
        <div class="pt-1 px-4">
          <div class="sticky top-0 z-10 h-0">
            <div class="flex justify-end">
              <button class="p-1 rounded text-xs z-text3 inline-flex items-center space-x-1 cursor-pointer" on:click={onCopy}>
                {#if !copied}
                  <ZIcon size={14} path={mdiContentCopy} />
                  <span>Copy</span>
                {:else}
                  <ZIcon size={14} path={mdiCheck} />
                  <span>Copied</span>
                {/if}
              </button>
            </div>
          </div>

          <div class="text-xs z-text3 select-none flex items-center gap-2">
            <span>{box.lang}</span>
            {#if jobValue.boxes.length > 1}
              <span>
                {#each jobValue.boxes as boxItem, i (i)}
                  <span title={boxItem.lang || ''} class={i !== seq ? 'opacity-30' : ''}>●</span>
                {/each}
              </span>
            {/if}
          </div>

          <div class="py-3">
            <!-- eslint-disable-next-line svelte/no-at-html-tags -->
            {@html renderedContentHtml}
          </div>
        </div>
      </svelte:component>
    </div>
  </div>
{/if}

<style>
  :global(.z-color-preview-swatch) {
    display: inline-block;
    width: 0.7em;
    height: 0.7em;
    border: 1px solid rgb(148 163 184 / 0.8);
    border-radius: 0.1rem;
    margin-right: 0.2rem;
    vertical-align: middle;
  }
</style>
