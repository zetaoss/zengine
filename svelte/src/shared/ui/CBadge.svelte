<svelte:options
  customElement={{
    props: {
      class: { type: "String" },
      href: { type: "String" },
      variant: { type: "String" },
    },
  }}
/>

<script lang="ts" module>
  import type { Snippet } from 'svelte'

  export type BadgeVariant = 'default' | 'secondary' | 'destructive' | 'outline' | 'ghost' | 'link'

  const variantClasses: Record<BadgeVariant, string> = {
    default: 'cn-badge-variant-default',
    secondary: 'cn-badge-variant-secondary',
    destructive: 'cn-badge-variant-destructive',
    outline: 'cn-badge-variant-outline',
    ghost: 'cn-badge-variant-ghost',
    link: 'cn-badge-variant-link',
  }

  export function badgeVariants(opts: { variant?: BadgeVariant }): string {
    return [
      'cn-badge inline-flex w-fit shrink-0 items-center justify-center overflow-hidden whitespace-nowrap transition-colors focus-visible:ring-[3px] focus-visible:border-ring focus-visible:ring-ring/50 aria-invalid:ring-destructive/20 aria-invalid:border-destructive group/badge [&>svg]:pointer-events-none',
      'rounded-full',
      variantClasses[opts.variant ?? 'default'],
    ]
      .filter(Boolean)
      .join(' ')
  }

  export type BadgeProps = {
    class?: string
    ref?: HTMLElement | null
    href?: string
    variant?: BadgeVariant
    children?: Snippet
  }
</script>

<script lang="ts">
  let {
  class: className,
  ref = $bindable(null),
  href = undefined,
  variant = 'default',
  children,
}: BadgeProps = $props()

  let classes = $derived(badgeVariants({ variant }))
</script>

<svelte:element
  this={href ? 'a' : 'span'}
  bind:this={ref}
  data-slot="badge"
  {href}
  class={classes + (className ? ` ${className}` : '')}
>
{#if children}
    {@render children?.()}
  {/if}
</svelte:element>
