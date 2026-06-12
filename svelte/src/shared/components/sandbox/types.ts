import type { ConsoleLogLevel } from '../../utils/console'

export interface SandboxLog {
  level: ConsoleLogLevel
  args: unknown[]
}
