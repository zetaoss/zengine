<!-- BoxApex.svelte -->
<script lang="ts">
  import { mdiCheck, mdiContentCopy } from '@mdi/js'
  import type { Writable } from 'svelte/store'

  import ZIcon from '$shared/ui/ZIcon.svelte'
  import { colorizeCss, colorizeHtml } from '$shared/utils/colorize'

  import BoxFront from './BoxFront.svelte'
  import BoxLang from './BoxLang.svelte'
  import BoxNotebook from './BoxNotebook.svelte'
  import BoxZero from './BoxZero.svelte'
  import type { Job } from './types'

  export let job: Writable<Job>
  export let seq: number
  export let content = ''

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
  $: rendered = box.lang === 'css' ? colorizeCss(content) : box.lang === 'html' ? colorizeHtml(content) : content

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
            {@html rendered}
          </div>
        </div>
      </svelte:component>
    </div>
  </div>
{/if}

<style>
  :global(.colorize) {
    display: inline-block;
    width: 0.7em;
    height: 0.7em;
    border: 1px solid rgb(148 163 184 / 0.8);
    border-radius: 0.1rem;
    margin-right: 0.2rem;
    vertical-align: middle;
  }
</style>
