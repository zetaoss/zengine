import { describe, expect, it } from 'vitest'

import { maskEmail } from './mask'

describe('maskEmail', () => {
  it('keeps first two characters and masks the rest in local and domain segments', () => {
    expect(maskEmail('john@example.com')).toBe('jo**@ex*****.co*')
  })

  it('does not trim input', () => {
    expect(maskEmail('  john@example.com  ')).toBe('  ****@ex*****.co***')
    expect(maskEmail('john@example.com')).toBe('jo**@ex*****.co*')
  })

  it('handles short inputs and multi-dot forms', () => {
    expect(maskEmail('a@b.c')).toBe('a@b.c')
    expect(maskEmail('ab@xy.com')).toBe('ab@xy.co*')
    expect(maskEmail('john@sub.example.com')).toBe('jo**@su*.ex*****.co*')
    expect(maskEmail('john.sub@example.com')).toBe('jo**.su*@ex*****.co*')
    expect(maskEmail('john@sub@gmail@com')).toBe('jo**@su*@gm***@co*')
    expect(maskEmail('john.sub.example.com')).toBe('jo**.su*.ex*****.co*')
  })

  it('does not throw on invalid email-like inputs', () => {
    expect(maskEmail('plain-text')).toBe('pl********')
    expect(maskEmail('abc@')).toBe('ab*@')
    expect(maskEmail('abc@localhost')).toBe('ab*@lo*******')
    expect(maskEmail('@abc.com')).toBe('@ab*.co*')
    expect(maskEmail('@')).toBe('@')
    expect(maskEmail('')).toBe('')
  })
})
