export type ConsoleLogLevel = 'log' | 'error' | 'warn' | 'info' | 'debug' | 'trace'

export const CONSOLE_EMOJIS: Partial<Record<ConsoleLogLevel, string>> = {
  debug: '🐛',
  error: '❌',
  info: 'ℹ️',
  trace: '🔍',
  warn: '⚠️',
}

export function getConsoleEmoji(level: ConsoleLogLevel | string): string {
  return CONSOLE_EMOJIS[level as ConsoleLogLevel] ?? ''
}
