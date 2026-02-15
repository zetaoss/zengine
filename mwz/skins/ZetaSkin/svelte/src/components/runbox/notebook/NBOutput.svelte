<!-- NBOutput.svelte -->
<script lang="ts">
  import './ansi.css'
  import './dataframe.css'

  import type { Output } from '../types'

  export let out: Output
  const ansiSgrRegex = new RegExp(`${String.fromCharCode(27)}\\[([\\d;]*)m`, 'g')

  const ansi2html = (input: string): string => {
    let open = false
    const escaped = input.replace(/&/g, '&amp;').replace(/</g, '&lt;')
    const converted = escaped.replace(ansiSgrRegex, (_match, codesStr: string) => {
      const codes = codesStr ? codesStr.split(';').map((c) => parseInt(c, 10) || 0) : [0]
      let html = ''
      for (const code of codes) {
        if (code === 0) {
          if (open) {
            html += '</span>'
            open = false
          }
          continue
        }
        if (open) html += '</span>'
        html += `<span class="ansi${code}">`
        open = true
      }
      return html
    })
    return converted + (open ? '</span>' : '')
  }

  $: textPlain = out.data?.['text/plain']?.join('') ?? ''
  $: textHtml = out.data?.['text/html']?.join('') ?? ''
  $: textLatex = out.data?.['text/latex']?.join('') ?? ''
  $: textJson = out.data?.['application/json']?.join('') ?? ''
</script>

<div
  class={`bg-[var(--console-bg)] px-4 py-2 text-sm font-mono rounded-lg ${out.output_type === 'error' ? 'bg-[var(--nbout-error)]' : ''}`}
>
  {#if out.output_type === 'display_data'}
    {#if out.data?.['image/png']}
      <img src={`data:image/png;base64,${out.data['image/png']}`} alt="" class="bg-white my-2 rounded max-h-[30rem] max-w-[75vw]" />
    {:else if textHtml}
      <!-- eslint-disable-next-line svelte/no-at-html-tags -->
      {@html textHtml}
    {:else if textPlain}
      <pre class="whitespace-pre-wrap">{textPlain}</pre>
    {:else if textJson}
      <pre class="bg-gray-100 dark:bg-gray-900 p-2 rounded overflow-auto">{textJson}</pre>
    {:else if textLatex}
      <div class="text-purple-700 dark:text-purple-300">\({textLatex}\)</div>
    {/if}
  {:else if out.output_type === 'execute_result'}
    {#if textHtml}
      <!-- eslint-disable-next-line svelte/no-at-html-tags -->
      {@html textHtml}
    {:else if textPlain}
      <pre class="whitespace-pre-wrap">{textPlain}</pre>
    {:else if textLatex}
      <div class="text-purple-700 dark:text-purple-300">\({textLatex}\)</div>
    {:else if textJson}
      <pre class="bg-gray-100 dark:bg-gray-900 p-2 rounded overflow-auto">{textJson}</pre>
    {/if}
  {:else if out.output_type === 'stream'}
    <pre class="whitespace-pre-wrap">{out.text?.join('')}</pre>
  {:else if out.output_type === 'error'}
    <!-- eslint-disable-next-line svelte/no-at-html-tags -->
    <pre>{@html ansi2html(out.traceback?.join('\n') ?? '')}</pre>
  {:else}
    <pre class="text-red-500">[Unsupported output_type: {out.output_type}]</pre>
  {/if}
</div>
