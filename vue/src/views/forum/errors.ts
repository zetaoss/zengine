// errors.ts
export class DataErrors {
  private errors: Map<string, Set<string>>

  constructor() {
    this.errors = new Map()
  }

  add(field: string, message: string) {
    if (!this.errors.has(field)) {
      this.errors.set(field, new Set())
    }
    this.errors.get(field)?.add(message)
  }

  get(field: string): string[] {
    return Array.from(this.errors.get(field) || [])
  }

  has(field: string): boolean {
    return this.errors.has(field) && this.errors.get(field)!.size > 0
  }

  isError(): boolean {
    return this.errors.size > 0
  }

  clear(field: string) {
    this.errors.delete(field)
  }

  clearAll() {
    this.errors.clear()
  }
}

export interface ErrorResponse {
  data: {
    error: Record<string, string[]>
  }
}

export function useErrors() {
  return new DataErrors()
}
