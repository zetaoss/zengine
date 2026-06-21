import { writable } from 'svelte/store'

export interface SelectedAiEditResult {
  content: string
  taskId: number
}

export const selectedAiEditResult = writable<SelectedAiEditResult | null>(null)
