import { describe, expect, it } from 'vitest'

import { maskEmail } from './mask'

describe('maskEmail', () => {
  it('keeps first two characters and masks the rest in local, host and tld', () => {
    expect(maskEmail('john@example.com')).toBe('jo**@ex*****.co*')
  })

  it('does not trim input', () => {
    expect(maskEmail('  john@gmail.com  ')).toBe('  ****@gm***.co***')
    expect(maskEmail('john@gmail.com')).toBe('jo**@gm***.co*')
  })

  it('handles short inputs', () => {
    expect(maskEmail('a@b.c')).toBe('a@b.c')
    expect(maskEmail('ab@xy.com')).toBe('ab@xy.co*')
  })

  it('does not throw on invalid email-like inputs', () => {
    expect(maskEmail('plain-text')).toBe('pl********@.')
    expect(maskEmail('abc@')).toBe('ab*@.')
    expect(maskEmail('abc@localhost')).toBe('ab*@lo*******.')
    expect(maskEmail('@abc.com')).toBe('@ab*.co*')
    expect(maskEmail('@')).toBe('@.')
    expect(maskEmail('')).toBe('@.')
  })
})
