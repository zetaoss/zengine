// @common/types/user.ts
export interface User {
  id: number
  name: string
}

// https://www.mediawiki.org/wiki/API:Userinfo
export interface UserInfo extends User {
  groups: string[]
  rights: string[]
}
