<script lang="ts">
  const PREVIEW_COUNT = 5

  export let value: unknown

  type Kind = 'null' | 'undefined' | 'Array' | 'Boolean' | 'Function' | 'Number' | 'Object' | 'String' | 'Other'

  type Entry = [string, unknown]

  function getKind(val: unknown): Kind {
    if (val === null) return 'null'
    if (val === undefined) return 'undefined'

    const t = typeof val
    if (t === 'string') return 'String'
    if (t === 'number') return 'Number'
    if (t === 'boolean') return 'Boolean'
    if (Array.isArray(val)) return 'Array'
    if (t === 'function') return 'Function'
    if (t === 'object') return 'Object'

    return 'Other'
  }

  function getEntries(val: unknown, k: Kind): Entry[] {
    if (k === 'Array') {
      return (val as unknown[]).map((v, i) => [String(i), v])
    }

    if (k === 'Object') {
      return Object.entries(val as Record<string, unknown>)
    }

    return []
  }

  function getDisplayName(val: unknown, k: Kind): string {
    if (k === 'Array') {
      return `Array(${(val as unknown[]).length})`
    }

    if (k === 'Object') {
      const ctorName = (val as { constructor?: { name?: string } })?.constructor?.name
      return ctorName || 'Object'
    }

    return k
  }

  function formatValue(v: unknown): string {
    const k = getKind(v)

    switch (k) {
      case 'Array':
      case 'Object':
        return getDisplayName(v, k)
      case 'String':
        return `'${String(v)}'`
      case 'Function':
        return String(v).replace('function', 'ƒ')
      default:
        return String(v)
    }
  }

  function getClassFromKind(k: Kind): string {
    return k.toLowerCase()
  }

  $: kind = getKind(value)
  $: isSingle =
    kind === 'null' || kind === 'undefined' || kind === 'String' || kind === 'Number' || kind === 'Boolean' || kind === 'Function'
  $: isArray = kind === 'Array'
  $: isObject = kind === 'Object'

  $: entries = getEntries(value, kind)
  $: preview = entries.slice(0, PREVIEW_COUNT)
  $: hasMore = entries.length > PREVIEW_COUNT

  let isOpen = false
</script>

{#if isSingle}
  <span class={getClassFromKind(kind)}>{formatValue(value)}</span>
{:else}
  <details class="inline-block align-top" bind:open={isOpen}>
    <summary>
      {#if isArray}
        {#if isOpen}
          <span>{getDisplayName(value, kind)}</span>
        {:else}
          <span>({(value as unknown[]).length}) </span>
          {#if preview.length}
            <span>[</span>
            {#each preview as [k, v], idx (k)}
              {#if idx > 0}
                <!-- eslint-disable-next-line svelte/no-useless-mustaches -->
                <span>,{' '}</span>
              {/if}
              <span class={getClassFromKind(getKind(v))}>{formatValue(v)}</span>
            {/each}
            {#if hasMore}
              <!-- eslint-disable-next-line svelte/no-useless-mustaches -->
              <span>,{' '}…</span>
            {/if}
            <span>]</span>
          {/if}
        {/if}
      {:else if isObject}
        <span>{getDisplayName(value, kind)}</span>
        {#if preview.length}
          <!-- eslint-disable-next-line svelte/no-useless-mustaches -->
          <span>{'{'}</span>
          {#each preview as [k, v], idx (k)}
            {#if idx > 0}
              <!-- eslint-disable-next-line svelte/no-useless-mustaches -->
              <span>,{' '}</span>
            {/if}
            <span class="objkey">{k}</span>
            <!-- eslint-disable-next-line svelte/no-useless-mustaches -->
            <span>:{' '}</span>
            <span class={getClassFromKind(getKind(v))}>{formatValue(v)}</span>
          {/each}
          {#if hasMore}
            <!-- eslint-disable-next-line svelte/no-useless-mustaches -->
            <span>,{' '}…</span>
          {/if}
          <!-- eslint-disable-next-line svelte/no-useless-mustaches -->
          <span>{'}'}</span>
        {/if}
      {/if}
    </summary>

    {#if isOpen}
      <ul class="list-none pl-4">
        {#each entries as [key, val] (key)}
          <li class="m-0">
            <span class={isArray ? 'arrkey' : 'objkey'}>{key}: </span>
            <svelte:self value={val} />
          </li>
        {/each}
      </ul>
    {/if}
  </details>
{/if}
