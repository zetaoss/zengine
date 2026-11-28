<svelte:options runes={true} />

<script lang="ts">
  import { onMount } from 'svelte'

  import { mdiChevronDown } from '@mdi/js'

  import { showToast } from '$shared/ui/toast/toast'
  import ZIcon from '$shared/ui/ZIcon.svelte'
  import httpy from '$shared/utils/httpy'
  import { getAge } from '$shared/utils/time'

  import { fetchModelIds, formatModelLabel } from '../loop'

  // A step's messages as written, a jinja2 template per role (prompts.MessageTemplates).
  interface MessageTemplates {
    system?: string
    user?: string
  }

  const messageRoles = ['system', 'user'] as const

  // A step: a jinja2 template per message, read in its output variable's format (prompts.Step).
  interface ChainStep {
    name: string
    description?: string
    message: MessageTemplates
    output_variable?: { name?: string; format?: string }
    inputs: string[]
  }

  // A variable the task running the chain fills (prompts.Input).
  interface ChainInput {
    name: string
    description?: string
  }

  interface Chain {
    name: string
    description: string
    version: string
    inputs: ChainInput[]
    steps: ChainStep[]
    model: string
    uses: number
    used_at?: string | null
  }

  let chains = $state<Chain[]>([])
  let modelIds = $state<string[]>([])
  let loading = $state(true)
  let savingChain = $state('')
  let expanded = $state<string | null>(null)
  const now = Date.now()

  async function fetchChains() {
    const [data, err] = await httpy.get<Chain[]>('/api/ai/chains')
    loading = false
    if (err) {
      showToast(err.message || '체인을 불러오지 못했습니다.')
      return
    }
    // Most recently used first; chains never run keep their order (that of goapp/prompts) after them.
    const usedAt = (c: Chain) => (c.used_at ? new Date(c.used_at).getTime() : 0)
    chains = (data ?? [])
      .map((c, i) => ({ c, i }))
      .sort((a, b) => usedAt(b.c) - usedAt(a.c) || a.i - b.i)
      .map(({ c }) => c)
  }

  async function setModel(chain: Chain, model: string) {
    savingChain = chain.name
    const [data, err] = await httpy.put<{ model: string }>(`/api/ai/chains/${chain.name}/model`, { model })
    savingChain = ''
    if (err || !data) {
      showToast(err?.message || '모델 저장에 실패했습니다.')
      return
    }
    chain.model = data.model
    showToast('저장했습니다')
  }

  function modelOptions(chain: Chain): string[] {
    return modelIds.includes(chain.model) ? modelIds : [chain.model, ...modelIds]
  }

  function toggle(name: string) {
    expanded = expanded === name ? null : name
  }

  onMount(() => {
    fetchChains()
    fetchModelIds().then((ids) => (modelIds = ids))
  })
</script>

{#snippet expandButton(key: string, isExpanded: boolean)}
  <button
    type="button"
    class="text-muted-foreground hover:text-foreground transition p-1 inline-flex items-center justify-center rounded hover:bg-muted"
    aria-label={isExpanded ? '접기' : '펼치기'}
    onclick={(e) => {
      e.stopPropagation()
      toggle(key)
    }}
  >
    <ZIcon path={mdiChevronDown} size={18} class={`transition-transform duration-200 ${isExpanded ? 'rotate-180' : ''}`} />
  </button>
{/snippet}

{#snippet stepView(st: ChainStep, n: number)}
  <div class="space-y-1.5 text-xs min-w-0">
    <div>
      <span class="text-muted-foreground">{n + 1}.</span>
      <span class="font-mono font-medium">{st.name}</span>
      <span class="ml-1 font-mono text-muted-foreground"
        >{st.inputs.join(', ')} → {st.output_variable?.format ?? 'text'}{st.output_variable?.name
          ? ` → ${st.output_variable.name}`
          : ''}</span
      >
      <span class="ml-2 text-muted-foreground">{st.description}</span>
    </div>
    {#each messageRoles.filter((role) => st.message[role]) as role (role)}
      <div class="relative overflow-hidden rounded border bg-muted/30">
        <span class="pointer-events-none absolute -top-px right-[13px] rounded px-1.5 py-0.5 font-mono text-muted-foreground/70"
          >{role}</span
        >
        <pre class="max-h-96 overflow-auto whitespace-pre-wrap p-3 pr-16 text-[11px]">{st.message[role]}</pre>
      </div>
    {/each}
  </div>
{/snippet}

<div class="container mx-auto p-4 md:p-6 max-w-6xl space-y-6">
  <h1 class="text-2xl font-bold tracking-tight border-b pb-4">체인</h1>

  <div class="rounded-lg border bg-card overflow-hidden shadow-sm">
    <div class="overflow-x-auto">
      <table class="w-full text-left border-collapse text-sm">
        <thead>
          <tr class="border-b bg-muted/50 text-muted-foreground text-xs font-medium">
            <th class="p-3">체인</th>
            <th class="p-3">설명</th>
            <th class="p-3">모델</th>
            <th class="p-3 text-center whitespace-nowrap">사용</th>
            <th class="p-3 whitespace-nowrap">최근</th>
            <th class="p-3 w-14 text-center">상세</th>
          </tr>
        </thead>
        <tbody class="divide-y">
          {#if loading}
            <tr><td colspan="6" class="p-8 text-center text-muted-foreground">불러오는 중...</td></tr>
          {/if}
          {#each chains as chain (chain.name)}
            {@const isExpanded = expanded === chain.name}
            <tr class="hover:bg-muted/30 transition cursor-pointer" onclick={() => toggle(chain.name)}>
              <td class="p-3 whitespace-nowrap">{chain.name}</td>
              <td class="p-3 text-muted-foreground">{chain.description}</td>
              <td class="p-3 text-xs">{formatModelLabel(chain.model)}</td>
              <td class="p-3 text-center text-muted-foreground tabular-nums" title="이 체인이 끝까지 돈 횟수">{chain.uses || '-'}</td>
              <td class="p-3 whitespace-nowrap text-xs text-muted-foreground">{chain.used_at ? getAge(chain.used_at, now) : '-'}</td>
              <td class="p-3 text-center">
                {@render expandButton(chain.name, isExpanded)}
              </td>
            </tr>
            {#if isExpanded}
              <tr class="bg-muted/10">
                <td colspan="6" class="p-4">
                  <div class="space-y-4">
                    <div class="flex flex-wrap items-center gap-2 text-xs">
                      <label class="flex min-w-0 flex-1 items-center gap-2">
                        <span class="font-medium">모델</span>
                        <select
                          class="h-8 w-full max-w-72 rounded border bg-background px-2.5 text-xs focus:outline-none focus:ring-1 focus:ring-primary"
                          value={chain.model}
                          disabled={savingChain === chain.name}
                          onchange={(e) => setModel(chain, e.currentTarget.value)}
                        >
                          {#each modelOptions(chain) as id (id)}
                            <option value={id}>{formatModelLabel(id)}</option>
                          {/each}
                        </select>
                      </label>
                    </div>
                    <div class="space-y-1 text-xs text-muted-foreground">
                      <div>버전: <span class="font-mono">{chain.version}</span></div>
                      <div>task 입력: <span class="font-mono">{chain.inputs.map((i) => i.name).join(', ')}</span></div>
                    </div>
                    {#each chain.steps as st, n (st.name)}
                      {@render stepView(st, n)}
                    {/each}
                  </div>
                </td>
              </tr>
            {/if}
          {/each}
        </tbody>
      </table>
    </div>
  </div>
</div>
