export type BinderItem = {
  id?: number
  href?: string
  new?: boolean | number
  nodes?: BinderItem[]
  text: string
}

export type Binder = {
  id: number
  title: string
  trees: BinderItem[]
}
