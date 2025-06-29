import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import Spinner from '../Spinner.vue'

describe('Spinner.vue', () => {
  it('renders with default props', () => {
    const wrapper = mount(Spinner);
    const svg = wrapper.find('svg');

    expect(svg.exists()).toBe(true);
    expect(svg.attributes('aria-hidden')).toBe('true');
    expect(svg.attributes('class')).toContain('animate-spin');
    expect(svg.attributes('style')).toContain('width: 2rem');
    expect(svg.attributes('style')).toContain('height: 2rem');
  });

  it('accepts a custom size', () => {
    const wrapper = mount(Spinner, {
      props: { size: '3rem' }
    });
    expect(wrapper.find('svg').attributes('style')).toContain('width: 3rem');
    expect(wrapper.find('svg').attributes('style')).toContain('height: 3rem');
  });

  it('accepts extraClass prop', () => {
    const wrapper = mount(Spinner, {
      props: { extraClass: 'custom-class' }
    });
    expect(wrapper.find('svg').attributes('class')).toContain('custom-class');
  });
});
