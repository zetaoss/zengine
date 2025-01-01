export interface Binder {
  id: number
  text: string
  title: string
  trees: BinderNodeType[]
}
export interface BinderNodeType {
  href: string
  text: string
  new: boolean
  nodes: BinderNodeType[]
}
