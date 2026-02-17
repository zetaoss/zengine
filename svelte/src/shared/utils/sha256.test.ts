import { afterEach, describe, expect, it, vi } from 'vitest'

import { sha256Hex } from './sha256'

describe('sha256Hex', () => {
  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('returns SHA-256 hex and null error', async () => {
    const [hash, err] = await sha256Hex('test@example.com')

    expect(err).toBeNull()
    expect(hash).toBe('973dfe463ec85785f5f95af5ba3906eedb2d931c24e69824a89ea65dba4e813b')
  })

  it('returns error tuple when crypto digest throws', async () => {
    const digestSpy = vi.spyOn(crypto.subtle, 'digest').mockRejectedValueOnce(new Error('boom'))

    const [hash, err] = await sha256Hex('test@example.com')

    expect(digestSpy).toHaveBeenCalledOnce()
    expect(hash).toBe('')
    expect(err).toBeInstanceOf(Error)
    expect(err?.message).toBe('boom')
  })
})
