// @/types/binder.ts

export interface Binder {
  id: number
  text: string
  title: string
  trees: BinderNodeData[]
}

export interface BinderNodeData {
  id: number
  text: string
  new: boolean
  href?: string
  nodes?: BinderNodeData[]
}
