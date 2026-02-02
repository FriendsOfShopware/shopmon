import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import { defineComponent, h } from 'vue';
import Login from './Login.vue';

// Stubs
const AlertStub = defineComponent({
  name: 'Alert',
  props: ['type', 'dismissible'],
  template: '<div :class="`alert-${type}`"><slot /></div>',
});

const RouterLinkStub = defineComponent({
  name: 'RouterLink',
  props: ['to'],
  setup(props, { slots }) {
    return () => h('a', { href: JSON.stringify(props.to) }, slots.default?.());
  },
});

const FieldStub = defineComponent({
  name: 'Field',
  props: ['name', 'type', 'placeholder', 'class', 'id'],
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () => h('input', {
      name: props.name,
      type: props.type || 'text',
      placeholder: props.placeholder,
      class: props.class,
      id: props.id,
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
const mockPush = vi.fn();
vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}));

// Mock auth client
vi.mock('@/helpers/auth-client', () => ({
  authClient: {
    signIn: {
      email: vi.fn(),
      passkey: vi.fn(),
      social: vi.fn(),
      sso: vi.fn(),
    },
    useSession: vi.fn(() => ({
      data: { user: { id: '1', email: 'test@example.com' } },
    })),
  },
}));

// Mock useAlert composable
vi.mock('@/composables/useAlert', () => ({
  useAlert: () => ({
    error: vi.fn(),
    success: vi.fn(),
  }),
}));

// Mock useReturnUrl composable
vi.mock('@/composables/useReturnUrl', () => ({
  useReturnUrl: () => ({
    returnUrl: { value: null },
    clearReturnUrl: vi.fn(),
  }),
}));

describe('Login', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  function mountComponent() {
    return mount(Login, {
      global: {
        stubs: {
          Alert: AlertStub,
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

  it('displays sign in heading', () => {
    const wrapper = mountComponent();
    expect(wrapper.find('h2').text()).toBe('Sign in to your account');
  });

  it('displays link to create account', () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain('New to Shopmon?');
    expect(wrapper.text()).toContain('Create an account');
  });

  it('displays database migration notice', () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain('Database Migration Notice');
    expect(wrapper.text()).toContain('you may need to reset your password');
  });

  it('has email input field', () => {
    const wrapper = mountComponent();
    const emailInput = wrapper.find('input[name="email"]');
    expect(emailInput.exists()).toBe(true);
    expect(emailInput.attributes('placeholder')).toBe('Email address');
  });

  it('has password input field', () => {
    const wrapper = mountComponent();
    const passwordInput = wrapper.find('input[name="password"]');
    expect(passwordInput.exists()).toBe(true);
    expect(passwordInput.attributes('placeholder')).toBe('Password');
  });

  it('has sign in button', () => {
    const wrapper = mountComponent();
    const signInButton = wrapper.find('button[type="submit"]');
    expect(signInButton.exists()).toBe(true);
    expect(signInButton.text()).toContain('Sign in');
  });

  it('has forgot password link', () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain('Forgot your password?');
  });

  it('has passkey login button', () => {
    const wrapper = mountComponent();
    const passkeyButton = wrapper.findAll('button').find(b => b.text().includes('Passkey'));
    expect(passkeyButton).toBeTruthy();
  });

  it('has GitHub login button', () => {
    const wrapper = mountComponent();
    const githubButton = wrapper.findAll('button').find(b => b.text().includes('GitHub'));
    expect(githubButton).toBeTruthy();
  });

  it('has SSO section with email input', () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain('Enterprise SSO');
    expect(wrapper.text()).toContain('Use your company email to sign in with SSO');
    
    const ssoInput = wrapper.findAll('input').find(i => i.attributes('placeholder')?.includes('work email'));
    expect(ssoInput).toBeTruthy();
  });

  it('has form element', () => {
    const wrapper = mountComponent();
    const form = wrapper.find('form');
    expect(form.exists()).toBe(true);
  });

  it('displays divider between login methods', () => {
    const wrapper = mountComponent();
    expect(wrapper.text()).toContain('Or');
  });
});
