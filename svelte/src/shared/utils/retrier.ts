export function useRetrier(
  task: () => void,
  initialDelay = 1000,
  maxDelay = 30000,
  factor = 1.1,
  maxAttempts = 0,
) {
  let delay = initialDelay
  let timeoutId: ReturnType<typeof setTimeout> | null = null
  let attempts = 0

  const executeTask = () => {
    attempts += 1
    task()
  }

  function schedule(): void {
    if (timeoutId) return
    if (maxAttempts > 0 && attempts >= maxAttempts) {
      clear()
      return
    }

    timeoutId = setTimeout(() => {
      timeoutId = null
      delay = Math.min(delay * factor, maxDelay)
      executeTask()
    }, delay)
  }

  function start(): void {
    clear()
    delay = initialDelay
    attempts = 0
    executeTask()
  }

  function clear() {
    if (timeoutId) {
      clearTimeout(timeoutId)
      timeoutId = null
    }
    delay = initialDelay
    attempts = 0
  }

  return { schedule, start, clear }
}
