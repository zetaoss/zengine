export type Link = {
  accesskey?: string
  active?: boolean
  class?: string | false
  context?: string
  current?: boolean
  exists?: boolean
  href: string
  icon?: string
  id?: string
  primary?: boolean
  redundant?: boolean
  text: string
  title?: string
  'data-mw'?: string
  'link-class'?: string
  'single-id'?: string
}

export type LinkMap = Record<string, Link>
export type LinkMapMap = Record<string, LinkMap>
