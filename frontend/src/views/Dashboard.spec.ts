import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { defineComponent, h, ref } from 'vue';
import Dashboard from './Dashboard.vue';

// Stubs
const HeaderContainerStub = defineComponent({
  name: 'HeaderContainer',
  props: ['title'],
  template: '<header>{{ title }}</header>',
});

const StatusIconStub = defineComponent({
  name: 'StatusIcon',
  props: ['status'],
  template: '<span :class="status">Status</span>',
});

const DataTableStub = defineComponent({
  name: 'DataTable',
  props: ['columns', 'data'],
  template: '<div class="data-table"><slot v-for="item in data" :row="item" /></div>',
});

const RouterLinkStub = defineComponent({
  name: 'RouterLink',
  props: ['to'],
  setup(props, { slots }) {
    return () => h('a', { href: JSON.stringify(props.to) }, slots.default?.());
  },
});

// Mock data
const mockShops = [
  {
    id: '1',
    name: 'Test Shop 1',
    organizationSlug: 'test-org',
    projectName: 'Test Project',
    shopwareVersion: '6.5.0',
    status: 'green',
    favicon: null,
  },
  {
    id: '2',
    name: 'Test Shop 2',
    organizationSlug: 'test-org',
    projectName: 'Another Project',
    shopwareVersion: '6.4.0',
    status: 'red',
    favicon: '/favicon.ico',
  },
];

const mockOrganizations = [
  {
    id: '1',
    name: 'Test Organization',
    slug: 'test-org',
    memberCount: 5,
    shopCount: 2,
  },
];

const mockChangelogs = [
  {
    id: '1',
    shopId: '1',
    shopName: 'Test Shop 1',
    organizationSlug: 'test-org',
    extensions: [{ name: 'Test Extension', oldVersion: '1.0.0', newVersion: '1.1.0' }],
    date: new Date('2024-01-15').toISOString(),
  },
];

// Mock trpcClient
const mockQueryResults: Record<string, any> = {
  'account.listOrganizations': Promise.resolve(mockOrganizations),
  'account.currentUserShops': Promise.resolve(mockShops),
  'account.currentUserChangelogs': Promise.resolve(mockChangelogs),
};

vi.mock('@/helpers/trpc', () => ({
  trpcClient: {
    account: {
      listOrganizations: {
        query: () => mockQueryResults['account.listOrganizations'],
      },
      currentUserShops: {
        query: () => mockQueryResults['account.currentUserShops'],
      },
      currentUserChangelogs: {
        query: () => mockQueryResults['account.currentUserChangelogs'],
      },
    },
  },
}));

// Mock changelog helper
vi.mock('@/helpers/changelog', () => ({
  sumChanges: (row: any) => row.extensions?.length || 0,
}));

// Mock formatter
vi.mock('@/helpers/formatter', () => ({
  formatDateTime: (date: string) => new Date(date).toLocaleString(),
}));

describe('Dashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset mock data
    mockQueryResults['account.listOrganizations'] = Promise.resolve(mockOrganizations);
    mockQueryResults['account.currentUserShops'] = Promise.resolve(mockShops);
    mockQueryResults['account.currentUserChangelogs'] = Promise.resolve(mockChangelogs);
  });

  function mountComponent() {
    return mount(Dashboard, {
      global: {
        stubs: {
          HeaderContainer: HeaderContainerStub,
          StatusIcon: StatusIconStub,
          DataTable: DataTableStub,
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

  it('displays dashboard title', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('header').text()).toBe('Dashboard');
  });

  it('displays My Shops section', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain('My Shops');
  });

  it('displays shops data', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain('Test Shop 1');
    expect(wrapper.text()).toContain('Test Shop 2');
    expect(wrapper.text()).toContain('Test Project');
    expect(wrapper.text()).toContain('6.5.0');
  });

  it('displays shop status icons', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const statusIcons = wrapper.findAll('.green, .red');
    expect(statusIcons.length).toBe(2);
  });

  it('displays My Organizations section', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain('My Organizations');
  });

  it('displays organizations data', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain('Test Organization');
    expect(wrapper.text()).toContain('5 Members');
    expect(wrapper.text()).toContain('2 Shops');
  });

  it('displays Last Changes section when changelogs exist', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain('Last Changes');
  });

  it('does not display Last Changes section when no changelogs', async () => {
    mockQueryResults['account.currentUserChangelogs'] = Promise.resolve([]);
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).not.toContain('Last Changes');
  });

  it('displays changelog data', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain('Test Shop 1');
  });

  it('displays correct shop links', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    const links = wrapper.findAll('a');
    const shopLink = links.find(l => l.text().includes('Test Shop 1'));
    expect(shopLink).toBeTruthy();
  });
});
