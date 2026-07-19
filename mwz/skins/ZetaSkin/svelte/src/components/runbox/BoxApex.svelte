<!-- BoxApex.svelte -->
<script lang="ts">
  import { mdiCheck, mdiContentCopy, mdiDotsVertical, mdiReplay, mdiWrap, mdiWrapDisabled } from '@mdi/js'
  import type { Writable } from 'svelte/store'

  import getRLCONF from '$lib/utils/rlconf'
  import CButton from '$shared/ui/CButton.svelte'
  import CMenu from '$shared/ui/CMenu.svelte'
  import CMenuItem from '$shared/ui/CMenuItem.svelte'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import { colorizeCss, colorizeHtml } from '$shared/utils/colorize'

  import BoxFront from './BoxFront.svelte'
  import BoxLang from './BoxLang.svelte'
  import BoxNotebook from './BoxNotebook.svelte'
  import BoxZero from './BoxZero.svelte'
  import { rerunJob } from './runbox'
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
  $: isMain = seq === jobValue.main
  const isSysop = (getRLCONF().wgUserGroups || []).includes('sysop')

  let copied = false
  let copyTimer: number | null = null
  let wrapped = true

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
    <div class="bg-a-gray-50 border rounded-lg pt-0.5">
      <svelte:component this={CurrentComponent} {job} {seq} {wrapped}>
        <div class="sticky top-0 z-10 h-0">
          <div class="flex justify-end items-center pr-0.5">
            <CButton variant="ghost" size="sm" onclick={onCopy}>
              {#if !copied}
                <ZIcon path={mdiContentCopy} />
                <span>Copy</span>
              {:else}
                <ZIcon path={mdiCheck} />
                <span>Copied</span>
              {/if}
            </CButton>
            <CMenu>
              {#snippet trigger({ toggle })}
                <CButton variant="ghost" size="sm" aria-label="Runbox menu" onclick={toggle}>
                  <ZIcon path={mdiDotsVertical} />
                </CButton>
              {/snippet}
              {#snippet menu({ close })}
                <CMenuItem
                  onclick={() => {
                    wrapped = !wrapped
                    close()
                  }}
                >
                  <ZIcon size={14} path={wrapped ? mdiWrapDisabled : mdiWrap} />
                  <span>{wrapped ? 'Unwrap' : 'Wrap'}</span>
                </CMenuItem>
                {#if isSysop && (isMain || jobValue.type === 'notebook') && (jobValue.phase === 'pending' || jobValue.phase === 'failed' || jobValue.phase === 'succeeded')}
                  <CMenuItem
                    onclick={() => {
                      void rerunJob(job)
                      close()
                    }}
                  >
                    <ZIcon size={14} path={mdiReplay} />
                    <span>Rerun</span>
                  </CMenuItem>
                {/if}
              {/snippet}
            </CMenu>
          </div>
        </div>

        <div class="text-xs text-muted-foreground select-none flex items-center gap-2">
          <span class="pl-4">{box.lang}</span>
          {#if jobValue.boxes.length > 1}
            <span>
              {#each jobValue.boxes as boxItem, i (i)}
                <span title={boxItem.lang || ''} class={i !== seq ? 'opacity-30' : ''}>●</span>
              {/each}
            </span>
          {/if}
        </div>

        <div class:code-wrapped={wrapped} class:code-unwrapped={!wrapped} class="px-4 py-3">
          <!-- eslint-disable-next-line svelte/no-at-html-tags -->
          {@html rendered}
        </div>
      </svelte:component>
    </div>
  </div>
{/if}

<style>
  .code-wrapped {
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  .code-wrapped :global(pre) {
    white-space: inherit;
    overflow-wrap: inherit;
  }

  .code-unwrapped {
    display: block;
    width: 100%;
    min-width: 0;
    max-width: 100%;
    overflow-x: auto;
    white-space: pre;
    word-break: normal;
    overflow-wrap: normal;
  }

  .code-unwrapped :global(*) {
    white-space: inherit !important;
    word-break: inherit !important;
    overflow-wrap: inherit !important;
  }

  .code-unwrapped :global(pre) {
    width: max-content;
    min-width: 100%;
  }

  :global(.colorize) {
    display: inline-block;
    width: 0.7em;
    height: 0.7em;
    border: 1px solid rgb(148 163 184 / 0.8);
    border-radius: 0.1rem;
    margin: 0.2rem;
    vertical-align: middle;
  }
</style>
