export async function sha256Hex(input: string): Promise<[string, Error | null]> {
  try {
    const data = new TextEncoder().encode(input)
    const digest = await crypto.subtle.digest('SHA-256', data)
    const bytes = new Uint8Array(digest)

    let out = ''
    for (const b of bytes) {
      out += b.toString(16).padStart(2, '0')
    }

    return [out, null]
  } catch (err: unknown) {
    if (err instanceof Error) {
      return ['', err]
    }

    return ['', new Error('Failed to create SHA-256 hash')]
  }
}
