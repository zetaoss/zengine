<script lang="ts">
  export let pill = true
  export let text = ''
  export let tone: 'neutral' | 'destructive' = 'neutral'

  $: classes = [
    'inline-flex shrink-0 items-center justify-center border px-1 text-xs font-medium leading-none',
    pill ? 'rounded-full' : 'rounded',
    tone === 'destructive'
      ? 'bg-[var(--background-color-destructive-subtle,#fee7e6)] text-[var(--color-destructive)]'
      : 'bg-[var(--background-color-interactive-subtle)] text-[var(--color-subtle)]',
    typeof $$props.class === 'string' ? $$props.class : '',
  ]
    .filter(Boolean)
    .join(' ')
</script>

<span {...$$restProps} class={classes}>
  {text}
</span>
