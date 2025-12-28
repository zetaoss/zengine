// mask.test.ts
import { describe, expect, it } from 'vitest'

import { maskEmail } from './mask'

describe('maskEmail', () => {
  it.each([
    ['abc@example.com', 'a*c@e*****e.c*m'],
    ['a@b.c', 'a@b.c'],
    ['ab@cd.ef', 'ab@cd.ef'],
    ['', ''],
  ])('%s → %s', (input, expected) => {
    expect(maskEmail(input)).toBe(expected)
  })
})
