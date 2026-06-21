import type { Config } from 'svelte/compiler'

export function isCustomElement(options: { filename: string }): boolean

declare const config: Config
export default config
