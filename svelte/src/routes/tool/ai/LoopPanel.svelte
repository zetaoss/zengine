<svelte:options runes={true} />

<script lang="ts">
  // The heading, settings (model, interval, strategy), Instant button and run status (last, running, next) of
  // one loop, shared by the definition and evaluation tabs.
  import { displayTitle } from '$shared/utils/mediawiki'
  import { getAge } from '$shared/utils/time'
  import { getWikiViewHref } from '$shared/utils/wikiLink'

  import {
    formatCountdown,
    formatModelLabel,
    formatTime,
    intervalOptionsWith,
    type LastRun,
    type LoopSettings,
    type QueuedTask,
    queueStateLabels,
    strategyLabels,
    strategyOptions,
  } from './loop'

  let {
    heading,
    instantTitle,
    settings,
    modelIds,
    saving,
    lastRun,
    tasks,
    now,
    instantDisabled,
    onchange,
    oninstant,
  }: {
    heading: string
    instantTitle: string
    settings: LoopSettings
    modelIds: string[]
    saving: boolean
    lastRun: LastRun | null
    tasks: QueuedTask[]
    now: number
    instantDisabled: boolean
    onchange: (change: Partial<LoopSettings>) => void
    oninstant: () => void
  } = $props()

  let modelOptions = $derived(settings.model && !modelIds.includes(settings.model) ? [settings.model, ...modelIds] : modelIds)
  let running = $derived(tasks.filter((t) => t.state === 'active'))
  let waiting = $derived(tasks.filter((t) => t.state === 'pending' || t.state === 'scheduled'))
  let strategyLabel = $derived(strategyLabels[settings.strategy] ?? '랜덤')

  function pageHref(pageID: number | undefined, title: string): string {
    return (pageID ?? 0) > 0 ? `/w/index.php?curid=${pageID}` : getWikiViewHref(title)
  }

  const selectClass = 'h-8 rounded border bg-background px-2.5 text-xs focus:outline-none focus:ring-1 focus:ring-primary'
</script>

{#snippet pageCell(task: QueuedTask | undefined)}
  {#if task?.page_title}
    <a href={pageHref(task.page_id, task.page_title)} target="_blank" rel="noopener">{displayTitle(task.page_title)}</a>
  {:else if task}
    <span class="text-muted-foreground">{strategyLabel}</span>
  {:else}
    <span class="text-muted-foreground/60">-</span>
  {/if}
{/snippet}

<div class="flex flex-wrap items-center justify-between gap-4 border-b pb-4">
  <h1 class="text-2xl font-bold tracking-tight">{heading}</h1>
  <div class="flex flex-wrap items-center gap-2">
    <select
      class={`${selectClass} max-w-64`}
      value={settings.model}
      disabled={saving}
      aria-label="모델"
      title="모델"
      onchange={(e) => onchange({ model: e.currentTarget.value })}
    >
      {#each modelOptions as id (id)}
        <option value={id}>{formatModelLabel(id)}</option>
      {/each}
    </select>
    <select
      class={`${selectClass} w-24`}
      value={settings.interval}
      disabled={saving}
      aria-label="주기"
      title="주기"
      onchange={(e) => onchange({ interval: e.currentTarget.value })}
    >
      {#each intervalOptionsWith(settings.interval) as option (option.value)}
        <option value={option.value}>{option.label}</option>
      {/each}
    </select>
    <select
      class={`${selectClass} w-24`}
      value={settings.strategy}
      disabled={saving}
      aria-label="자동선정 방식"
      title="자동선정 방식"
      onchange={(e) => onchange({ strategy: e.currentTarget.value })}
    >
      {#each strategyOptions as option (option.value)}
        <option value={option.value}>{option.label}</option>
      {/each}
    </select>
    <button
      type="button"
      class="h-8 rounded border bg-background px-3 text-xs font-medium text-foreground transition hover:bg-muted disabled:opacity-50 cursor-pointer disabled:cursor-not-allowed"
      disabled={instantDisabled || running.length > 0}
      title={instantTitle}
      onclick={oninstant}
    >
      Instant
    </button>
  </div>
</div>

<div class="flex flex-wrap items-start justify-between gap-4 rounded-lg border bg-muted/30 p-3 text-xs text-muted-foreground">
  <div class="space-y-0.5 min-w-0">
    <div class="font-medium text-foreground">
      최근 {#if lastRun}<span class="font-normal text-muted-foreground">{getAge(lastRun.at, now)}</span>{:else}<span
          class="font-normal text-muted-foreground/60">-</span
        >{/if}
    </div>
    <div class="truncate">
      {#if lastRun?.page_title}
        <a href={pageHref(lastRun.page_id, lastRun.page_title)} target="_blank" rel="noopener">{displayTitle(lastRun.page_title)}</a>
      {:else}
        <span class="text-muted-foreground/60">-</span>
      {/if}
    </div>
    <div class="truncate">
      {#if lastRun?.model}
        <span>{formatModelLabel(lastRun.model)}</span>
      {:else}
        <span class="text-muted-foreground/60">-</span>
      {/if}
      {#if lastRun}
        {#if lastRun.ok}
          <span class="ml-1 font-medium text-a-emerald-600" title="성공">✓</span>
        {:else}
          <span class="ml-1 font-medium text-a-rose-600" title="실패">✗</span>
        {/if}
      {/if}
    </div>
    {#if lastRun && !lastRun.ok && lastRun.error}
      <div class="text-a-rose-500 break-words">{lastRun.error}</div>
    {/if}
  </div>

  {#if running.length > 0}
    <div class="space-y-0.5 min-w-0">
      <div class="font-medium text-foreground flex items-center gap-1.5">
        <span>실행 중</span>
        <span class="h-2 w-2 rounded-full bg-a-blue-500 animate-pulse"></span>
      </div>
      <div class="truncate">{@render pageCell(running[0])}</div>
      <div class="truncate"><span>{formatModelLabel(settings.model)}</span></div>
    </div>
  {/if}

  <div class="space-y-0.5 min-w-0">
    <div class="font-medium text-foreground">
      다음
      {#if waiting.length === 0}
        <span class="font-normal text-muted-foreground">{settings.interval === '0s' ? '꺼짐' : '없음'}</span>
      {:else if waiting[0].next_process_at}
        <span class="font-normal text-muted-foreground" title={`${formatTime(waiting[0].next_process_at)} 예정`}
          >{formatCountdown(waiting[0].next_process_at, now)}</span
        >
      {:else}
        <span class="font-normal text-muted-foreground">{queueStateLabels[waiting[0].state]}</span>
      {/if}
    </div>
    <div class="truncate">{@render pageCell(waiting[0])}</div>
    <div class="truncate">
      {#if waiting.length > 0 && settings.interval !== '0s'}
        <span>{formatModelLabel(settings.model)}</span>
      {:else}
        <span class="text-muted-foreground/60">-</span>
      {/if}
    </div>
  </div>
</div>
