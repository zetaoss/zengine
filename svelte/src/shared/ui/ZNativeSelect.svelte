<svelte:options runes={true} />

<script lang="ts">
  interface SelectItem {
    group?: string
    value: string
    label: string
  }

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
    onchange?: (value: string) => void
  } = $props()

  let groupedItems = $derived(
    items.reduce<Array<{ group?: string; items: SelectItem[] }>>((acc, item) => {
      const lastGroup = acc[acc.length - 1]
      if (item.group && lastGroup?.group === item.group) {
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

  function handleChange(event: Event) {
    const select = event.currentTarget as HTMLSelectElement
    value = select.value
    onchange?.(select.value)
  }
</script>

<select bind:value class="z-select {className}" onchange={handleChange}>
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
