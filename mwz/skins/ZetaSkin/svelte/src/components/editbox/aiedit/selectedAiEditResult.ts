import { writable } from 'svelte/store'

import type { AIEditRequestType } from './aiEditTypes'

export interface SelectedAiEditResult {
  content: string
  taskId: number
  requestType: AIEditRequestType
  onRequestTypeChange?: (newType: AIEditRequestType) => void
}

export const selectedAiEditResult = writable<SelectedAiEditResult | null>(null)
