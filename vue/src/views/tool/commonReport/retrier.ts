export function useRetrier(
  task: () => void,
  initialDelay = 1000,
  maxDelay = 30000,
  factor = 1.1,
) {
  let delay = initialDelay
  let timeoutId: ReturnType<typeof setTimeout> | null = null

  const executeTask = () => {
    console.log(delay)
    task()
  }

  function schedule(): void {
    if (timeoutId || delay > maxDelay) return
    timeoutId = setTimeout(() => {
      executeTask()
      delay *= factor
      timeoutId = null
    }, delay)
  }

  function start(): void {
    clear()
    delay = initialDelay
    executeTask()
  }

  function clear() {
    if (timeoutId) {
      clearTimeout(timeoutId)
      timeoutId = null
    }
    delay = initialDelay
  }

  return { schedule, start, clear }
}
