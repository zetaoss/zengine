export interface ItemData {
  type?: string
  icon: string
  title?: string
  action?: () => {}
  isActive?: () => {}
}
