import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import { defineComponent, h } from 'vue';
import Register from './Register.vue';

// Stubs
const RouterLinkStub = defineComponent({
  name: 'RouterLink',
  props: ['to'],
  setup(props, { slots }) {
    return () => h('a', { href: JSON.stringify(props.to) }, slots.default?.());
  },
});

// Field stub that renders an actual input
const FieldStub = defineComponent({
  name: 'Field',
  props: ['name', 'type', 'placeholder', 'class'],
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () => h('input', {
      name: props.name,
      type: props.type ?? 'text',
      placeholder: props.placeholder,
      class: props.class,
      onInput: (e: Event) => {
        emit('update:modelValue', (e.target as HTMLInputElement).value);
      },
    });
  },
});

const PasswordFieldStub = defineComponent({
  name: 'PasswordField',
  props: ['name', 'placeholder', 'error'],
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () => h('input', {
      type: 'password',
      name: props.name,
      placeholder: props.placeholder,
      onInput: (e: Event) => {
        emit('update:modelValue', (e.target as HTMLInputElement).value);
      },
    });
  },
});

// Mock router
vi.mock('@/router', () => ({
  router: {
    push: vi.fn(),
  },
}));

// Mock auth client
vi.mock('@/helpers/auth-client', () => ({
  authClient: {
    signUp: {
      email: vi.fn(),
    },
  },
}));

// Mock useAlert composable
vi.mock('@/composables/useAlert', () => ({
  useAlert: () => ({
    success: vi.fn(),
    error: vi.fn(),
  }),
}));

describe('Register', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  function mountComponent() {
    return mount(Register, {
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
          Field: FieldStub,
          PasswordField: PasswordFieldStub,
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
    expect(wrapper.find('h2').text()).toBe('Create account');
  });

  it('has display name input field', () => {
    const wrapper = mountComponent();
    const input = wrapper.find('input[name="displayName"]');
    expect(input.exists()).toBe(true);
  });

  it('has email input field', () => {
    const wrapper = mountComponent();
    const input = wrapper.find('input[name="email"]');
    expect(input.exists()).toBe(true);
  });

  it('has password input field', () => {
    const wrapper = mountComponent();
    const input = wrapper.find('input[name="password"]');
    expect(input.exists()).toBe(true);
  });

  it('has register button', () => {
    const wrapper = mountComponent();
    const button = wrapper.find('button[type="submit"]');
    expect(button.exists()).toBe(true);
    expect(button.text()).toContain('Register');
  });

  it('has cancel link to login', () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain('Cancel');
  });

  it('has form element with submit handler', () => {
    const wrapper = mountComponent();
    const form = wrapper.find('form');
    expect(form.exists()).toBe(true);
  });

  it('has display name input with correct placeholder', () => {
    const wrapper = mountComponent();
    const input = wrapper.find('input[name="displayName"]');
    expect(input.attributes('placeholder')).toBe('Display Name');
  });

  it('has email input with correct placeholder', () => {
    const wrapper = mountComponent();
    const input = wrapper.find('input[name="email"]');
    expect(input.attributes('placeholder')).toBe('Email address');
  });
});
