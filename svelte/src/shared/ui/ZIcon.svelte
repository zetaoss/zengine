<script lang="ts">
  export let path: string
  export let size = 16
  let className = ''
  export { className as class }
</script>

<svg class={className} width={size} height={size} viewBox="0 0 24 24" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
  <path d={path} />
</svg>
