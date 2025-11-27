export type LogLevel = 'log' | 'error' | 'warn' | 'info' | 'debug' | 'trace'

export interface Log {
  level: LogLevel
  args: unknown[]
}
