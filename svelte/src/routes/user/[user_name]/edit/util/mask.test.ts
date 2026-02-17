import { describe, expect, it } from 'vitest'

import { maskEmail } from './mask'

describe('maskEmail', () => {
  it('preserves @ and keeps one of every two characters', () => {
    expect(maskEmail('test@example.com')).toBe('t*s*@*x*m*l*.*o*')
  })

  it('does not trim input', () => {
    expect(maskEmail('  test@example.com  ')).toBe(' *t*s*@*x*m*l*.*o* *')
  })

  it('handles short inputs', () => {
    expect(maskEmail('a@b.c')).toBe('a@b*c')
    expect(maskEmail('ab@xy.com')).toBe('a*@*y*c*m')
  })

  it('still masks when @ is missing', () => {
    expect(maskEmail('plain-text')).toBe('p*a*n*t*x*')
  })
})
