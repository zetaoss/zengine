<script lang="ts">
  import ZBadge from './ZBadge.svelte'

  interface ZTabItem {
    value: string
    label: string
    badge?: string | number
    href?: string
  }

  export let tabs: ZTabItem[] = []
  export let selected = ''
  export let onChange: ((value: string) => void) | undefined = undefined
  let pendingSelected = ''

  function beginSelect(event: PointerEvent, value: string) {
    if (event.button !== 0) return
    if (event.ctrlKey || event.metaKey || event.shiftKey) return
    if (value === selected) return
    if (onChange) {
      pendingSelected = value
    }
  }

  function beginKeyboardSelect(event: KeyboardEvent, value: string) {
    if (value === selected) return
    if (event.key !== 'Enter') return
    if (onChange) {
      pendingSelected = value
    }
  }

  function select(event: MouseEvent, value: string) {
    if (event.ctrlKey || event.metaKey || event.shiftKey) return
    if (value === selected) {
      event.preventDefault()
      return
    }
    if (onChange) {
      event.preventDefault()
      pendingSelected = value
      onChange(value)
    }
  }

  function resolve(href: string) {
    return href
  }

  $: if (pendingSelected && selected === pendingSelected) {
    pendingSelected = ''
  }
</script>

<div class="mb-4 inline-flex items-end">
  {#each tabs as tab (tab.value)}
    <a
      href={resolve(tab.href ?? '#')}
      class={`relative border-b-2 px-3 py-2 text-sm transition ${
        selected === tab.value || pendingSelected === tab.value
          ? 'border-slate-900 font-semibold text-slate-900'
          : 'border-transparent text-gray-500 hover:text-gray-800'
      }`}
      onpointerdown={(e) => beginSelect(e, tab.value)}
      onkeydown={(e) => beginKeyboardSelect(e, tab.value)}
      onclick={(e) => select(e, tab.value)}
    >
      <span class="inline-flex items-center gap-1">
        {tab.label}
        {#if tab.badge !== undefined}
          <ZBadge text={String(tab.badge)} />
        {/if}
      </span>
    </a>
  {/each}
</div>
