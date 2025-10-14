export interface Binder {
  id: number
  text: string
  title: string
  trees: BinderNodeType[]
}

export interface BinderNodeType {
  id: number
  href: string
  text: string
  new: boolean
  nodes: BinderNodeType[]
}
