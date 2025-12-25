// Split.test.ts
import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { nextTick } from 'vue'

import Split from '../Split.vue'

const resizeHandlers: Array<(entries: Array<{ contentRect: { width: number; height: number } }>) => void> = []

vi.mock('@vueuse/core', () => ({
  useEventListener: vi.fn(),
  useResizeObserver: vi.fn((_, cb: typeof resizeHandlers[number]) => {
    resizeHandlers.push(cb)
  }),
}))

describe('Split', () => {
  it('resizes vertically using the container width', async () => {
    resizeHandlers.length = 0

    const wrapper = mount(Split, {
      props: { direction: 'vertical', initialPercentage: 20 },
    })

    expect(resizeHandlers).toHaveLength(1)

    resizeHandlers[0]([{ contentRect: { width: 1000, height: 200 } }])
    await nextTick()

    const firstPane = wrapper.findAll('.overflow-auto')[0]
    expect(firstPane.attributes('style')).toContain('0 0 20%')
    expect(firstPane.attributes('style')).not.toContain('0 0 50%')
  })
})
