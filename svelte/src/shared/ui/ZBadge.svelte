<script lang="ts">
  export let pill = true

  $: classes = [
    'inline-flex items-center justify-center bg-[#9993] border px-1 text-xs font-medium leading-none',
    pill ? 'rounded-full' : 'rounded',
    typeof $$props.class === 'string' ? $$props.class : '',
  ]
    .filter(Boolean)
    .join(' ')
</script>

<span class={classes}>
  <slot />
</span>
