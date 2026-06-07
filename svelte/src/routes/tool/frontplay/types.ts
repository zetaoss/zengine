import type { ConsoleLogLevel } from '$shared/utils/console'

export interface SandboxLog {
  level: ConsoleLogLevel
  args: unknown[]
}
