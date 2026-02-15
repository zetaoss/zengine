<script lang="ts">
  export let n: number | undefined = undefined

  const grades = ['D', 'C', 'B', 'A', 'S']

  $: starCount = Math.max(0, (n ?? 0) + 1)
  $: grade = grades[n ?? 0] ?? ''
</script>

<span>{'★'.repeat(starCount)} {grade}</span>
