// avatar.ts
export interface Avatar {
  id: number
  name: string
  t: number // 3: gravatar, 2: letter, else: identicon
  ghash: string // gravatar hash
}