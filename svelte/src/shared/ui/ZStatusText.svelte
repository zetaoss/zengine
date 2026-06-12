<script lang="ts">
  export let text = ''

  $: className = typeof $$props.class === 'string' ? $$props.class : ''
</script>

<span {...$$restProps} class={`inline-flex items-center text-muted-foreground ${className}`.trim()} aria-label={text}>
  {text}
</span>
