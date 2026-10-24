<svelte:options runes={true} />

<script lang="ts">
  // Shows a bob model ID with its provider prefix (up to the first "/") dimmed, e.g. groq/openai/gpt-oss-120b.
  let { id }: { id: string } = $props()

  let slash = $derived(id.indexOf('/'))
</script>

{#if slash > 0}<span class="text-muted-foreground/60">{id.slice(0, slash + 1)}</span>{id.slice(slash + 1)}{:else}{id}{/if}
