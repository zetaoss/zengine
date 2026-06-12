<script lang="ts">
  export let width = '10rem'
  export let height = '0.75rem'
  export let extraClass = ''
</script>

<span
  aria-hidden="true"
  class={`inline-block align-middle animate-pulse rounded bg-x-gray-300/70 ${extraClass}`}
  style={`width:${width}; height:${height};`}
></span>
