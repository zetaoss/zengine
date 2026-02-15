<!-- ConsoleArg.svelte -->
<script lang="ts">
  import ConsoleArg from './ConsoleArg.svelte'

  export let value: unknown

  const PREVIEW_COUNT = 5

  const Kind = {
    Null: 'null',
    Undefined: 'undefined',
    String: 'String',
    Number: 'Number',
    Boolean: 'Boolean',
    Array: 'Array',
    Object: 'Object',
    Function: 'Function',
    Other: 'Other',
  } as const

  type Kind = (typeof Kind)[keyof typeof Kind]

  function getKind(val: unknown): Kind {
    if (val === null) return Kind.Null
    if (val === undefined) return Kind.Undefined

    const t = typeof val

    if (t === 'string') return Kind.String
    if (t === 'number') return Kind.Number
    if (t === 'boolean') return Kind.Boolean
    if (Array.isArray(val)) return Kind.Array
    if (t === 'function') return Kind.Function
    if (t === 'object') return Kind.Object

    return Kind.Other
  }

  const getEntries = (val: unknown, k: Kind): [string, unknown][] => {
    if (k === Kind.Array) {
      return (val as unknown[]).map((v, i) => [String(i), v])
    }

    if (k === Kind.Object) {
      return Object.entries(val as Record<string, unknown>)
    }

    return []
  }

  const getDisplayName = (val: unknown, k: Kind): string => {
    if (k === Kind.Array) {
      return `Array(${(val as unknown[]).length})`
    }

    if (k === Kind.Object) {
      const ctorName = (val as { constructor?: { name?: string } })?.constructor?.name
      return ctorName || Kind.Object
    }

    return k
  }

  const formatValue = (v: unknown): string => {
    const k = getKind(v)

    switch (k) {
      case Kind.Array:
      case Kind.Object:
        return getDisplayName(v, k)
      case Kind.String:
        return `'${v}'`
      case Kind.Function:
        return String(v)
      default:
        return String(v)
    }
  }

  const getClassFromKind = (k: Kind): string => k.toLowerCase()

  let isOpen = false

  $: kind = getKind(value)
  $: isSingle =
    kind === Kind.Null ||
    kind === Kind.Undefined ||
    kind === Kind.String ||
    kind === Kind.Number ||
    kind === Kind.Boolean ||
    kind === Kind.Function

  $: isArray = kind === Kind.Array
  $: isObject = kind === Kind.Object
  $: arrayLength = isArray ? (value as unknown[]).length : 0

  $: entries = getEntries(value, kind)
  $: preview = entries.slice(0, PREVIEW_COUNT)
  $: hasMore = entries.length > PREVIEW_COUNT
</script>

{#if isSingle}
  <span>
    <span class={getClassFromKind(kind)}>{formatValue(value)}</span>
  </span>
{:else}
  <details class="inline-block align-top" bind:open={isOpen}>
    <summary>
      {#if isArray}
        {#if isOpen}
          {getDisplayName(value, kind)}
        {:else}
          ({arrayLength})
          {#if preview.length}
            <span>[</span>
            {#each preview as entry, idx (entry[0])}
              <span class={getClassFromKind(getKind(entry[1]))}>{formatValue(entry[1])}</span>
              {#if idx < preview.length - 1}
                <span>, </span>
              {/if}
            {/each}
            {#if hasMore}
              <span>, ...</span>
            {/if}
            <span>]</span>
          {/if}
        {/if}
      {:else if isObject}
        {getDisplayName(value, kind)}
        {#if preview.length}
          <span>&#123;</span>
          {#each preview as entry, idx (entry[0])}
            <span class="objkey">{entry[0]}</span>: {formatValue(entry[1])}
            {#if idx < preview.length - 1}
              <span>, </span>
            {/if}
          {/each}
          {#if hasMore}
            <span>, ...</span>
          {/if}
          <span>&#125;</span>
        {/if}
      {/if}
    </summary>

    {#if isOpen}
      <ul class="list-none pl-4">
        {#each entries as entry (entry[0])}
          <li class="m-0">
            <span class={isArray ? 'arrkey' : 'objkey'}>{entry[0]}: </span>
            <ConsoleArg value={entry[1]} />
          </li>
        {/each}
      </ul>
    {/if}
  </details>
{/if}

<style>
  details > summary {
    cursor: pointer;
  }
</style>
