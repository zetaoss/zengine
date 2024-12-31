import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import TheSpinner from '../TheSpinner.vue'

describe('TheSpinner', () => {
  it('renders properly', () => {
    const wrapper = mount(TheSpinner)
    expect(wrapper.text()).toContain('Loading...')
  })
})
