export type LogLevel = 'log' | 'error' | 'warn' | 'info' | 'debug' | 'trace'

export interface SandboxLog {
  level: LogLevel
  args: unknown[]
}
