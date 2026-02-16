type MenuItem = {
  accesskey?: string
  active?: boolean
  class?: string | false
  exists?: boolean
  href: string
  icon?: string
  id?: string
  text: string
  title?: string
  'data-mw'?: string
  'link-class'?: string[]
  'single-id'?: string
}

export type MyMenu = Record<string, MenuItem>
export type PageMenu = MenuItem[]
