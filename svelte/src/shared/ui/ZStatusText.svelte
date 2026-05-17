<script lang="ts">
  export let text = ''

  $: className = typeof $$props.class === 'string' ? $$props.class : ''
</script>

<span {...$$restProps} class={`inline-flex items-center text-(--color-subtle) ${className}`.trim()} aria-label={text}>
  {text}
</span>
