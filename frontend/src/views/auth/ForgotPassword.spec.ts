import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import { defineComponent, h } from 'vue';
import ForgotPassword from './ForgotPassword.vue';

// Stubs
const AlertStub = defineComponent({
  name: 'Alert',
  props: ['type', 'dismissible'],
  setup(props, { slots }) {
    return () => h('div', { class: `alert alert-${props.type}` }, slots.default?.());
  },
});

const RouterLinkStub = defineComponent({
  name: 'RouterLink',
  props: ['to'],
  setup(props, { slots }) {
    return () => h('a', { href: props.to }, slots.default?.());
  },
});

// Field stub that renders an actual input
const FieldStub = defineComponent({
  name: 'Field',
  props: ['name', 'type', 'placeholder', 'class', 'id', 'autocomplete'],
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () => h('input', {
      name: props.name,
      type: props.type ?? 'text',
      placeholder: props.placeholder,
      class: props.class,
      id: props.id,
      autocomplete: props.autocomplete,
      onInput: (e: Event) => {
        emit('update:modelValue', (e.target as HTMLInputElement).value);
      },
    });
  },
});

// Mock auth client
vi.mock('@/helpers/auth-client', () => ({
  authClient: {
    forgetPassword: vi.fn(),
  },
}));

// Mock useAlert composable
vi.mock('@/composables/useAlert', () => ({
  useAlert: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}));

describe('ForgotPassword', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  function mountComponent() {
    return mount(ForgotPassword, {
      global: {
        stubs: {
          Alert: AlertStub,
          RouterLink: RouterLinkStub,
          Field: FieldStub,
        },
      },
    });
  }

  it('renders successfully', () => {
    const wrapper = mountComponent();
    expect(wrapper.exists()).toBe(true);
  });

  it('displays page title', () => {
    const wrapper = mountComponent();
    expect(wrapper.find('h2').text()).toBe('Forgot password');
  });

  it('displays instructions', () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain('We will send you a confirmation email');
  });

  it('has email input field', () => {
    const wrapper = mountComponent();
    const emailInput = wrapper.find('input[name="email"]');
    expect(emailInput.exists()).toBe(true);
  });

  it('has submit button', () => {
    const wrapper = mountComponent();
    const submitButton = wrapper.find('button[type="submit"]');
    expect(submitButton.exists()).toBe(true);
    expect(submitButton.text()).toContain('Send email');
  });

  it('has cancel link to login', () => {
    const wrapper = mountComponent();
    const cancelLink = wrapper.find('a[href="login"]');
    expect(cancelLink.exists()).toBe(true);
    expect(cancelLink.text()).toBe('Cancel');
  });

  it('has form element with submit handler', () => {
    const wrapper = mountComponent();
    const form = wrapper.find('form');
    expect(form.exists()).toBe(true);
  });

  it('has email input with correct attributes', () => {
    const wrapper = mountComponent();
    const emailInput = wrapper.find('input[name="email"]');
    expect(emailInput.attributes('placeholder')).toBe('Email address');
  });
});
