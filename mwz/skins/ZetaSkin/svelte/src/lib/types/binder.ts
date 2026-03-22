export type BinderNode = {
  id?: number
  href?: string
  new?: boolean | number
  nodes?: BinderNode[]
  text: string
}

export type Binder = BinderNode & {
  id: number
  nodes: BinderNode[]
}
