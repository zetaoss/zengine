<script lang="ts">
  export let text = ''
  export let animated = false
  export let mode: 'shimmer' | 'pulse' = 'shimmer'

  $: className = typeof $$props.class === 'string' ? $$props.class : ''
</script>

<span
  {...$$restProps}
  class={`inline-flex items-center ${animated ? `z-status-text is-${mode}` : ''} ${className}`.trim()}
  aria-label={text}
>
  {text}
</span>

<style>
  .z-status-text {
    color: var(--color-subtle);
  }

  .z-status-text.is-shimmer {
    animation: z-status-text-shimmer 2.6s linear infinite;
    -webkit-background-clip: text;
    background-clip: text;
    background-image: linear-gradient(
      110deg,
      var(--color-subtle) 0%,
      var(--color-subtle) 38%,
      var(--color-base) 50%,
      var(--color-subtle) 62%,
      var(--color-subtle) 100%
    );
    background-position: 120% 50%;
    background-size: 220% 100%;
    color: transparent;
  }

  .z-status-text.is-pulse {
    animation: z-status-text-pulse 1.8s ease-in-out infinite;
  }

  @keyframes z-status-text-shimmer {
    0%,
    15% {
      background-position: 120% 50%;
    }
    85%,
    100% {
      background-position: -20% 50%;
    }
  }

  @keyframes z-status-text-pulse {
    0%,
    100% {
      color: var(--color-subtle);
    }
    50% {
      color: var(--color-base);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .z-status-text.is-shimmer,
    .z-status-text.is-pulse {
      animation: none;
      background-image: none;
      color: var(--color-subtle);
    }
  }
</style>
