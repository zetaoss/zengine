function leftRotate(x: number, c: number): number {
  return (x << c) | (x >>> (32 - c))
}

function toHexLe(n: number): string {
  let out = ''
  for (let i = 0; i < 4; i++) {
    const b = (n >>> (i * 8)) & 0xff
    out += b.toString(16).padStart(2, '0')
  }
  return out
}

export function md5(input: string): string {
  const s: number[] = [
    7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22, 5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20, 4, 11, 16, 23, 4,
    11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23, 6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21,
  ]

  const k = new Array<number>(64)
  for (let i = 0; i < 64; i++) {
    k[i] = Math.floor(Math.abs(Math.sin(i + 1)) * 4294967296) >>> 0
  }

  const bytes = Array.from(new TextEncoder().encode(input))
  const bitLen = BigInt(bytes.length) * 8n

  bytes.push(0x80)
  while (bytes.length % 64 !== 56) {
    bytes.push(0)
  }

  for (let i = 0; i < 8; i++) {
    bytes.push(Number((bitLen >> BigInt(8 * i)) & 0xffn))
  }

  let a0 = 0x67452301
  let b0 = 0xefcdab89
  let c0 = 0x98badcfe
  let d0 = 0x10325476

  for (let offset = 0; offset < bytes.length; offset += 64) {
    const m = new Array<number>(16)
    for (let i = 0; i < 16; i++) {
      const j = offset + i * 4
      m[i] = (bytes[j] | (bytes[j + 1] << 8) | (bytes[j + 2] << 16) | (bytes[j + 3] << 24)) >>> 0
    }

    let a = a0
    let b = b0
    let c = c0
    let d = d0

    for (let i = 0; i < 64; i++) {
      let f: number
      let g: number

      if (i < 16) {
        f = (b & c) | (~b & d)
        g = i
      } else if (i < 32) {
        f = (d & b) | (~d & c)
        g = (5 * i + 1) % 16
      } else if (i < 48) {
        f = b ^ c ^ d
        g = (3 * i + 5) % 16
      } else {
        f = c ^ (b | ~d)
        g = (7 * i) % 16
      }

      const temp = d
      d = c
      c = b
      const sum = (a + f + k[i] + m[g]) >>> 0
      b = (b + leftRotate(sum, s[i])) >>> 0
      a = temp
    }

    a0 = (a0 + a) >>> 0
    b0 = (b0 + b) >>> 0
    c0 = (c0 + c) >>> 0
    d0 = (d0 + d) >>> 0
  }

  return `${toHexLe(a0)}${toHexLe(b0)}${toHexLe(c0)}${toHexLe(d0)}`
}
