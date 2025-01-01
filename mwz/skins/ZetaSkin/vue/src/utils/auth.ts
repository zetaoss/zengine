import getRLCONF from './rlconf'

export function isLoggedIn() {
  return getRLCONF().wgUserId > 0
}

export function isAdmin() {
  return getRLCONF().wgUserGroups.indexOf('sysop') !== -1
}

export function canWrite() {
  return isLoggedIn()
}

export function canEdit(id: number) {
  return isLoggedIn() && getRLCONF().avatar.id === id
}

export function canDelete(id: number) {
  return canEdit(id) || isAdmin()
}
