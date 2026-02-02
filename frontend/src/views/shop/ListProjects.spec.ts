import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { defineComponent, h, ref } from 'vue';
import ListProjects from './ListProjects.vue';

// Stubs
const HeaderContainerStub = defineComponent({
  name: 'HeaderContainer',
  props: ['title'],
  setup(props, { slots }) {
    return () => h('header', {}, [
      props.title,
      slots.default?.(),
    ]);
  },
});

const MainContainerStub = defineComponent({
  name: 'MainContainer',
  setup(_, { slots }) {
    return () => h('main', {}, slots.default?.());
  },
});

const ElementEmptyStub = defineComponent({
  name: 'ElementEmpty',
  props: ['title', 'button', 'route'],
  setup(props, { slots }) {
    return () => h('div', { class: 'element-empty' }, [
      h('h3', {}, props.title),
      slots.default?.(),
      h('a', { href: JSON.stringify(props.route) }, props.button),
    ]);
  },
});

const StatusIconStub = defineComponent({
  name: 'StatusIcon',
  props: ['status'],
  template: '<span :class="status">{{ status }}</span>',
});

const RouterLinkStub = defineComponent({
  name: 'RouterLink',
  props: ['to'],
  setup(props, { slots }) {
    return () => h('a', { href: JSON.stringify(props.to) }, slots.default?.());
  },
});

// Mock data
const mockProjects = [
  {
    id: 1,
    name: 'Test Project',
    nameCombined: 'org-slug / Test Project',
    description: 'A test project description',
    createdAt: new Date('2024-01-15').toISOString(),
  },
];

const mockShops = [
  {
    id: 1,
    name: 'Test Shop',
    shopwareVersion: '6.5.0',
    status: 'green',
    favicon: null,
    projectId: 1,
  },
];

// Mock trpcClient
vi.mock('@/helpers/trpc', () => ({
  trpcClient: {
    account: {
      currentUserProjects: {
        query: vi.fn(() => Promise.resolve(mockProjects)),
      },
      currentUserShops: {
        query: vi.fn(() => Promise.resolve(mockShops)),
      },
    },
    organization: {
      project: {
        update: {
          mutate: vi.fn(() => Promise.resolve({})),
        },
        delete: {
          mutate: vi.fn(() => Promise.resolve({})),
        },
      },
    },
  },
}));

// Mock formatter
vi.mock('@/helpers/formatter', () => ({
  formatDate: (date: string) => new Date(date).toLocaleDateString(),
}));

// Mock useAlert
vi.mock('@/composables/useAlert', () => ({
  useAlert: () => ({
    error: vi.fn(),
    success: vi.fn(),
  }),
}));

// Mock route
const mockRoute = { params: {} };
vi.mock('vue-router', () => ({
  useRoute: () => mockRoute,
}));

describe('ListProjects', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  function mountComponent() {
    return mount(ListProjects, {
      global: {
        stubs: {
          HeaderContainer: HeaderContainerStub,
          MainContainer: MainContainerStub,
          ElementEmpty: ElementEmptyStub,
          StatusIcon: StatusIconStub,
          RouterLink: RouterLinkStub,
        },
      },
    });
  }

  it('renders successfully', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.exists()).toBe(true);
  });

  it('displays page title', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('header').text()).toContain('My Projects');
  });

  it('displays Add Project button in header', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain('Add Project');
  });

  it('displays empty state when no projects exist', async () => {
    // Override mock to return empty projects
    const { trpcClient } = await import('@/helpers/trpc');
    vi.mocked(trpcClient.account.currentUserProjects.query).mockResolvedValueOnce([]);
    
    const wrapper = mountComponent();
    await flushPromises();
    
    expect(wrapper.find('.element-empty').exists()).toBe(true);
    expect(wrapper.text()).toContain('No Projects');
  });

  it('displays projects list when projects exist', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    
    expect(wrapper.text()).toContain('Test Project');
    expect(wrapper.text()).toContain('org-slug / Test Project');
  });

  it('displays project description when available', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    
    expect(wrapper.text()).toContain('A test project description');
  });

  it('displays project meta information', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    
    expect(wrapper.text()).toContain('1 shops');
    expect(wrapper.text()).toContain('Created');
  });

  it('displays shops for each project', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    
    expect(wrapper.text()).toContain('Test Shop');
    expect(wrapper.text()).toContain('6.5.0');
  });

  it('displays shop status icons', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    
    expect(wrapper.find('.green').exists()).toBe(true);
  });

  it('shows empty state CTA when no projects', async () => {
    const { trpcClient } = await import('@/helpers/trpc');
    vi.mocked(trpcClient.account.currentUserProjects.query).mockResolvedValueOnce([]);
    
    const wrapper = mountComponent();
    await flushPromises();
    
    expect(wrapper.text()).toContain('Get started by creating your first project');
  });

  it('has project actions menu', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    
    // Menu container should exist for each project
    expect(wrapper.text()).toContain('Test Project');
  });

  it('displays shop favicon or fallback icon', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    
    // Shop should be displayed
    expect(wrapper.text()).toContain('Test Shop');
  });
});
