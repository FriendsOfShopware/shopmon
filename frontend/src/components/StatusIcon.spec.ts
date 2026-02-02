import { describe, it, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import StatusIcon from './StatusIcon.vue';

describe('StatusIcon', () => {
  it('renders successfully with default props', () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: 'green',
      },
    });
    expect(wrapper.exists()).toBe(true);
  });

  it('renders with tooltip when tooltip prop is true', () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: 'green',
        tooltip: true,
      },
    });
    expect(wrapper.find('.has-tooltip').exists()).toBe(true);
    expect(wrapper.find('[data-tooltip="green"]').exists()).toBe(true);
  });

  it('renders without tooltip when tooltip prop is false', () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: 'green',
        tooltip: false,
      },
    });
    expect(wrapper.find('.has-tooltip').exists()).toBe(false);
  });

  it('applies icon-success class for green status', () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: 'green',
      },
    });
    expect(wrapper.find('.icon-success').exists()).toBe(true);
  });

  it('applies icon-error class for red status', () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: 'red',
      },
    });
    expect(wrapper.find('.icon-error').exists()).toBe(true);
  });

  it('applies icon-warning class for yellow status', () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: 'yellow',
      },
    });
    expect(wrapper.find('.icon-warning').exists()).toBe(true);
  });

  it('applies icon-muted class for inactive status', () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: 'inactive',
      },
    });
    expect(wrapper.find('.icon-muted').exists()).toBe(true);
  });

  it('applies icon-muted class for not installed status', () => {
    const wrapper = mount(StatusIcon, {
      props: {
        status: 'not installed',
      },
    });
    expect(wrapper.find('.icon-muted').exists()).toBe(true);
  });
});
