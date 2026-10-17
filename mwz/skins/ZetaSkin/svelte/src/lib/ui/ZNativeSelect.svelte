<svelte:options runes={true} />

<script lang="ts">
  interface SelectItem {
    group?: string
    value: string
    label: string
  }

  type ChangeHandler = (value: string) => void | Promise<void>

  let {
    value = $bindable(''),
    items = [],
    placeholder = '선택...',
    class: className = '',
    onchange,
  }: {
    value?: string
    items: SelectItem[]
    placeholder?: string
    class?: string
    onchange?: ChangeHandler
  } = $props()

  let selectElement: HTMLSelectElement | undefined

  let groupedItems = $derived(
    items.reduce<Array<{ group?: string; items: SelectItem[] }>>((acc, item) => {
      const lastGroup = acc[acc.length - 1]
      if (item.group && lastGroup?.group === item.group) {
        lastGroup.items.push(item)
        return acc
      }

      if (!item.group && lastGroup && !lastGroup.group) {
        lastGroup.items.push(item)
        return acc
      }

      if (item.group) {
        acc.push({ group: item.group, items: [item] })
        return acc
      }

      acc.push({ items: [item] })
      return acc
    }, []),
  )

  function restoreSelectViewport(scrollX: number, scrollY: number) {
    window.scrollTo(scrollX, scrollY)
    selectElement?.focus({ preventScroll: true })
  }

  async function handleChange(event: Event) {
    const select = event.currentTarget as HTMLSelectElement
    const scrollX = window.scrollX
    const scrollY = window.scrollY

    try {
      await onchange?.(select.value)
    } finally {
      restoreSelectViewport(scrollX, scrollY)
      requestAnimationFrame(() => restoreSelectViewport(scrollX, scrollY))
    }
  }
</script>

<select bind:this={selectElement} bind:value class="z-select {className}" onchange={handleChange}>
  <option value="" disabled={items.length > 0}>{placeholder}</option>

  {#each groupedItems as group, index (group.group ?? `group-${index}`)}
    {#if group.group}
      <optgroup label={group.group}>
        {#each group.items as item (item.value)}
          <option value={item.value}>{item.label}</option>
        {/each}
      </optgroup>
    {:else}
      {#each group.items as item (item.value)}
        <option value={item.value}>{item.label}</option>
      {/each}
    {/if}
  {/each}
</select>
