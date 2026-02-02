import { describe, it, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import Logo from './Logo.vue';

describe('Logo', () => {
  it('renders successfully', () => {
    const wrapper = mount(Logo);
    expect(wrapper.exists()).toBe(true);
  });

  it('renders an img element', () => {
    const wrapper = mount(Logo);
    expect(wrapper.find('img').exists()).toBe(true);
  });

  it('has correct src attribute', () => {
    const wrapper = mount(Logo);
    const img = wrapper.find('img');
    expect(img.attributes('src')).toBe('/shopmon-logo.svg');
  });

  it('has correct alt attribute', () => {
    const wrapper = mount(Logo);
    const img = wrapper.find('img');
    expect(img.attributes('alt')).toBe('Shopmon');
  });
});
