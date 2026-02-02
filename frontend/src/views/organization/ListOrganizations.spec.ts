import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { defineComponent, h, ref } from 'vue';
import ListOrganizations from './ListOrganizations.vue';

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
  props: ['title', 'route', 'button'],
  template: '<div class="element-empty"><h3>{{ title }}</h3><slot /><a :href="route.name">{{ button }}</a></div>',
});

const DataTableStub = defineComponent({
  name: 'DataTable',
  props: ['columns', 'data'],
  setup(props) {
    return () => h('table', { class: 'data-table' }, [
      h('thead', {}, h('tr', {}, props.columns.map((col: any) => 
        h('th', { key: col.key }, col.name)
      ))),
      h('tbody', {}, props.data.map((row: any) =>
        h('tr', { key: row.id }, props.columns.map((col: any) =>
          h('td', { key: col.key }, row[col.key])
        ))
      )),
    ]);
  },
});

const RouterLinkStub = defineComponent({
  name: 'RouterLink',
  props: ['to'],
  setup(props, { slots }) {
    return () => h('a', { href: JSON.stringify(props.to) }, slots.default?.());
  },
});

// Mock organizations data
const mockOrganizations = {
  data: [
    {
      id: '1',
      name: 'Test Organization',
      slug: 'test-org',
      memberCount: 5,
      shopCount: 2,
    },
    {
      id: '2',
      name: 'Another Org',
      slug: 'another-org',
      memberCount: 3,
      shopCount: 1,
    },
  ],
};

const emptyOrganizations = {
  data: [],
};

// Mock authClient
let mockOrganizationsData = ref(mockOrganizations);

vi.mock('@/helpers/auth-client', () => ({
  authClient: {
    useListOrganizations: () => mockOrganizationsData.value,
  },
}));

describe('ListOrganizations', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockOrganizationsData.value = mockOrganizations;
  });

  function mountComponent() {
    return mount(ListOrganizations, {
      global: {
        stubs: {
          HeaderContainer: HeaderContainerStub,
          MainContainer: MainContainerStub,
          ElementEmpty: ElementEmptyStub,
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

  it('displays page title', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('header').text()).toContain('My Organization');
  });

  it('displays Add Organization button in header', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain('Add Organization');
  });

  it('displays organizations in data table when data exists', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('.data-table').exists()).toBe(true);
    expect(wrapper.text()).toContain('Test Organization');
    expect(wrapper.text()).toContain('Another Org');
  });

  it('displays organization names in table', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain('Test Organization');
    expect(wrapper.text()).toContain('Another Org');
  });

  it('displays organization slugs', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain('test-org');
    expect(wrapper.text()).toContain('another-org');
  });

  it('displays empty state when no organizations exist', async () => {
    mockOrganizationsData.value = emptyOrganizations;
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.find('.element-empty').exists()).toBe(true);
    expect(wrapper.text()).toContain('No Organization');
  });

  it('shows add organization button in empty state', async () => {
    mockOrganizationsData.value = emptyOrganizations;
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain('Get started by adding your first organization');
    expect(wrapper.text()).toContain('Add Organization');
  });

  it('displays correct table columns', async () => {
    const wrapper = mountComponent();
    await flushPromises();
    expect(wrapper.text()).toContain('Name');
    expect(wrapper.text()).toContain('Slug');
  });
});
