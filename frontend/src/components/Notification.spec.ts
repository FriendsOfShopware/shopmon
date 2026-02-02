import { describe, it, expect, beforeEach, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { useAlert } from '@/composables/useAlert';
import Notification from './Notification.vue';
import { nextTick, ref, type Ref } from 'vue';

// Mock the useAlert composable
vi.mock('@/composables/useAlert', () => ({
  useAlert: vi.fn(),
}));

describe('Notification', () => {
  let alertRef: Ref<any>;

  beforeEach(() => {
    vi.clearAllMocks();
    alertRef = ref(null);
    (useAlert as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      alert: alertRef,
      success: vi.fn(),
      info: vi.fn(),
      error: vi.fn(),
      warning: vi.fn(),
      clear: vi.fn(),
    });
  });

  it('renders nothing when alert is null', () => {
    const wrapper = mount(Notification);
    expect(wrapper.find('.notification').exists()).toBe(false);
  });

  it('renders error notification when alert type is error', async () => {
    alertRef.value = {
      title: 'Error Title',
      message: 'Error message',
      type: 'error',
    };
    const wrapper = mount(Notification);
    await nextTick();
    expect(wrapper.find('.notification-error').exists()).toBe(true);
    expect(wrapper.find('.notification-title').text()).toBe('Error Title');
  });

  it('renders success notification when alert type is success', async () => {
    alertRef.value = {
      title: 'Success Title',
      message: 'Success message',
      type: 'success',
    };
    const wrapper = mount(Notification);
    await nextTick();
    expect(wrapper.find('.notification-success').exists()).toBe(true);
  });

  it('renders info notification when alert type is info', async () => {
    alertRef.value = {
      title: 'Info Title',
      message: 'Info message',
      type: 'info',
    };
    const wrapper = mount(Notification);
    await nextTick();
    expect(wrapper.find('.notification-info').exists()).toBe(true);
  });

  it('renders warning notification when alert type is warning', async () => {
    alertRef.value = {
      title: 'Warning Title',
      message: 'Warning message',
      type: 'warning',
    };
    const wrapper = mount(Notification);
    await nextTick();
    expect(wrapper.find('.notification-warning').exists()).toBe(true);
  });

  it('calls clear when close button is clicked', async () => {
    const mockClear = vi.fn();
    (useAlert as unknown as ReturnType<typeof vi.fn>).mockReturnValue({
      alert: alertRef,
      success: vi.fn(),
      info: vi.fn(),
      error: vi.fn(),
      warning: vi.fn(),
      clear: mockClear,
    });
    alertRef.value = {
      title: 'Test',
      message: 'Test message',
      type: 'info',
    };
    const wrapper = mount(Notification);
    await nextTick();
    const closeButton = wrapper.find('.notification-close');
    await closeButton.trigger('click');
    expect(mockClear).toHaveBeenCalled();
  });

  it('displays alert title and message', async () => {
    alertRef.value = {
      title: 'My Title',
      message: 'My detailed message',
      type: 'success',
    };
    const wrapper = mount(Notification);
    await nextTick();
    expect(wrapper.find('.notification-title').text()).toBe('My Title');
    expect(wrapper.text()).toContain('My detailed message');
  });
});
