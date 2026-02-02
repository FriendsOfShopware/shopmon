import { describe, it, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import Imprint from './Imprint.vue';

// Mock components
vi.mock('@/components/layout/HeaderContainer.vue', () => ({
  default: {
    name: 'HeaderContainer',
    props: ['title'],
    template: '<header>{{ title }}</header>',
  },
}));

vi.mock('@/components/layout/MainContainer.vue', () => ({
  default: {
    name: 'MainContainer',
    template: '<main><slot /></main>',
  },
}));

describe('Imprint', () => {
  it('renders successfully', () => {
    const wrapper = mount(Imprint);
    expect(wrapper.exists()).toBe(true);
  });

  it('displays the correct page title', () => {
    const wrapper = mount(Imprint);
    expect(wrapper.find('header').text()).toBe('Legal notice');
  });

  it('contains Contact section', () => {
    const wrapper = mount(Imprint);
    expect(wrapper.text()).toContain('Contact');
  });

  it('displays non-profit disclaimer', () => {
    const wrapper = mount(Imprint);
    expect(wrapper.text()).toContain('Non-Profit Website');
    expect(wrapper.text()).toContain("it's not required to provide a legal representative");
  });

  it('displays email contact information', () => {
    const wrapper = mount(Imprint);
    expect(wrapper.text()).toContain('E-Mail: shopmon at fos.gg');
  });

  it('uses main-container component', () => {
    const wrapper = mount(Imprint);
    expect(wrapper.find('main').exists()).toBe(true);
  });

  it('has correct CSS class on page container', () => {
    const wrapper = mount(Imprint);
    expect(wrapper.find('.imprint-page').exists()).toBe(true);
  });

  it('has correct CSS class on content container', () => {
    const wrapper = mount(Imprint);
    expect(wrapper.find('.imprint-content').exists()).toBe(true);
  });
});
