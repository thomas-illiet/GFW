import { mountSuspended } from '@nuxt/test-utils/runtime'
import { nextTick } from 'vue'
import { beforeAll, describe, expect, it } from 'vitest'

import AppEmptyState from '../app/components/App/AppEmptyState.vue'
import AppDeleteDialog from '../app/components/App/AppDeleteDialog.vue'
import AppFirewallMark from '../app/components/App/AppFirewallMark.vue'
import AppPageHeader from '../app/components/App/AppPageHeader.vue'
import AppRelativeTime from '../app/components/App/AppRelativeTime.vue'
import AppSectionCard from '../app/components/App/AppSectionCard.vue'
import DashboardActivityKpis from '../app/components/Dashboard/DashboardActivityKpis.vue'
import DashboardBreakdowns from '../app/components/Dashboard/DashboardBreakdowns.vue'
import DashboardRequestChart from '../app/components/Dashboard/DashboardRequestChart.vue'
import DashboardRequestKpis from '../app/components/Dashboard/DashboardRequestKpis.vue'
import DashboardWindowSelect from '../app/components/Dashboard/DashboardWindowSelect.vue'
import EcosystemDialog from '../app/components/Ecosystems/EcosystemDialog.vue'
import IngestionKeyDialog from '../app/components/Keys/IngestionKeyDialog.vue'
import IngestionKeyFilters from '../app/components/Keys/IngestionKeyFilters.vue'
import IngestionKeyRevokeDialog from '../app/components/Keys/IngestionKeyRevokeDialog.vue'
import LoginPanel from '../app/components/Login/LoginPanel.vue'
import LoginVisual from '../app/components/Login/LoginVisual.vue'
import MachineDialog from '../app/components/Machines/MachineDialog.vue'
import MachineFilters from '../app/components/Machines/MachineFilters.vue'
import MachinesTable from '../app/components/Machines/MachinesTable.vue'
import TrustedServiceDialog from '../app/components/Trusted/TrustedServiceDialog.vue'
import TrustedServicesTable from '../app/components/Trusted/TrustedServicesTable.vue'
import type { Dashboard, Ecosystem, TrustedService } from '../app/types/api'

beforeAll(() => {
  Object.defineProperty(globalThis, 'visualViewport', {
    configurable: true,
    value: {
      addEventListener: () => undefined,
      removeEventListener: () => undefined,
      width: 1280,
      height: 720,
      offsetLeft: 0,
      offsetTop: 0,
    },
  })
})

const ecosystem: Ecosystem = {
  ID: 'eco-1',
  Name: 'payments',
  DisplayName: 'Payments',
  IsActive: true,
  MachineCount: 2,
  CreatedAt: '2026-01-01T00:00:00Z',
  UpdatedAt: '2026-01-01T00:00:00Z',
}

const trustedService: TrustedService = {
  ID: 'trusted-1',
  Name: 'Core DNS',
  Description: 'Shared DNS resolvers',
  IsActive: true,
  Hostnames: [
    {
      ID: 'hostname-1',
      Hostname: 'dns.example.com',
      Status: 'fresh',
      IPs: [
        {
          ID: 'ip-1',
          IPv4: '192.0.2.10',
          FirstSeenAt: '2026-01-01T00:00:00Z',
          LastSeenAt: '2026-01-01T00:00:00Z',
          IsCurrent: true,
        },
      ],
    },
  ],
  CreatedAt: '2026-01-01T00:00:00Z',
  UpdatedAt: '2026-01-01T00:00:00Z',
}

const dashboard: Dashboard = {
  window: '24h',
  requests: {
    total: 12,
    previousTotal: 8,
    changePercent: 50,
    successRate: 91.7,
    errorRate: 8.3,
    notModified: 3,
    averageLatencyMs: 14,
    p95LatencyMs: 31,
  },
  activity: {
    active: 2,
    passive: 1,
    total: 3,
    activeRate: 66.7,
    machineActive: 1,
    machinePassive: 1,
    trustedActive: 1,
    trustedPassive: 0,
  },
  series: [{ at: '2026-01-01T10:00:00Z', requests: 12, errors: 1 }],
  ecosystems: [{ name: 'payments', requests: 12, uniqueIps: 2 }],
  formats: [{ name: 'json', count: 7 }],
  sources: [{ name: 'machine', count: 2 }],
}

describe('shared application components', () => {
  it('renders the brand, page header, section and empty state', async () => {
    const brand = await mountSuspended(AppFirewallMark, { props: { size: 40 } })
    expect(brand.find('[aria-label="Great Firewall"]').exists()).toBe(true)

    const header = await mountSuspended(AppPageHeader, {
      props: {
        title: 'Machines',
        copy: 'Inventory',
        icon: 'mdi-server-network',
        stat: '4',
        statLabel: 'Addresses',
      },
    })
    expect(header.text()).toContain('Machines')
    expect(header.text()).toContain('4')

    const section = await mountSuspended(AppSectionCard, {
      props: { title: 'Catalogue', subtitle: 'Trusted services' },
      slots: { default: 'content', actions: '<button>Add</button>' },
    })
    expect(section.text()).toContain('Catalogue')
    expect(section.text()).toContain('Add')

    const empty = await mountSuspended(AppEmptyState, {
      props: { icon: 'mdi-alert', title: 'Nothing here', text: 'Create one.' },
    })
    expect(empty.text()).toContain('Nothing here')
  })

  it('renders the dedicated login composition components', async () => {
    const visual = await mountSuspended(LoginVisual, {
      props: { statusLabel: 'Perimeter online', statusTone: 'online' },
    })
    expect(visual.text()).toContain('One wall.')
    expect(visual.find('img').attributes('src')).toBeTruthy()

    const panel = await mountSuspended(LoginPanel, {
      props: { pending: false, error: null },
    })
    expect(panel.text()).toContain('Keep the right routes open.')
    expect(panel.text()).toContain('Enter Great Firewall')
  })

  it('renders accessible relative timestamps', async () => {
    const timestamp = await mountSuspended(AppRelativeTime, {
      props: {
        value: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
        prefix: 'Synced',
      },
    })
    expect(timestamp.text()).toContain('Synced 5 days ago')
    expect(timestamp.find('.relative-time').attributes('title')).toBeTruthy()
  })
})

describe('dashboard components', () => {
  it('renders all KPI, chart and breakdown sections', async () => {
    expect(
      (
        await mountSuspended(DashboardRequestKpis, {
          props: { value: dashboard.requests, activity: dashboard.activity },
        })
      ).text(),
    ).toContain('Address coverage')
    expect(
      (
        await mountSuspended(DashboardActivityKpis, {
          props: { value: dashboard.activity },
        })
      ).text(),
    ).toContain('66.7%')
    expect(
      (
        await mountSuspended(DashboardRequestChart, {
          props: { items: dashboard.series },
        })
      ).text(),
    ).toContain('Request volume')
    expect(
      (
        await mountSuspended(DashboardBreakdowns, {
          props: { data: dashboard },
        })
      ).text(),
    ).toContain('payments')
  })

  it('renders meaningful dashboard empty states', async () => {
    const empty = { ...dashboard, ecosystems: [], formats: [], sources: [] }
    const wrapper = await mountSuspended(DashboardBreakdowns, {
      props: { data: empty },
    })
    expect(wrapper.text()).toContain('No ecosystem traffic')
    expect(wrapper.text()).toContain('No traffic mix yet')
  })

  it('shows request details when a traffic point is hovered', async () => {
    const wrapper = await mountSuspended(DashboardRequestChart, {
      props: {
        items: [
          ...dashboard.series,
          { at: '2026-01-01T11:00:00Z', requests: 2, errors: 0 },
        ],
      },
    })
    await wrapper.findAll('.chart-hit-area')[1]?.trigger('mouseenter')
    expect(wrapper.find('.chart-tooltip').text()).toContain('2 requests')
    expect(wrapper.find('.chart-tooltip').text()).toContain('0 errors')
    expect(wrapper.find('.chart-tooltip').classes()).not.toContain(
      'chart-tooltip--below',
    )
  })

  it('updates the selected dashboard window', async () => {
    const wrapper = await mountSuspended(DashboardWindowSelect, {
      props: { modelValue: '24h' },
    })
    expect(wrapper.text()).toContain('30 days')
  })
})

describe('inventory components', () => {
  it('renders filters and machine table states', async () => {
    const filters = await mountSuspended(MachineFilters, {
      props: {
        ecosystems: [ecosystem],
        search: '',
        ecosystem: '',
        source: '',
      },
    })
    expect(filters.text()).toContain('Search hostname or IP address')
    expect(filters.text()).toContain('All sources')

    const table = await mountSuspended(MachinesTable, {
      props: { items: [], loading: false, canEdit: true },
    })
    expect(table.text()).toContain('No machines found')
    expect(table.text()).toContain('Add machine')
  })

  it('renders ingestion key search and revoked visibility controls', async () => {
    const filters = await mountSuspended(IngestionKeyFilters, {
      props: { search: '', showRevoked: false },
    })
    expect(filters.text()).toContain('Search name or prefix')
    expect(filters.text()).toContain('Key status')
    expect(filters.text()).toContain('Active keys')
  })

  it('requires the exact ingestion key name before revocation', async () => {
    const wrapper = await mountSuspended(IngestionKeyRevokeDialog, {
      attachTo: document.body,
      props: {
        modelValue: true,
        loading: false,
        ingestionKey: {
          ID: 'key-1',
          Name: 'Demo discovery agent',
          Prefix: 'demo_discovery',
          CreatedAt: '2026-01-01T00:00:00Z',
        },
      },
    })
    const revokeButton = Array.from(
      document.body.querySelectorAll('button'),
    ).find((button) => button.textContent?.includes('Revoke key'))

    expect(document.body.textContent).toContain(
      'Type Demo discovery agent to confirm',
    )
    expect(revokeButton?.disabled).toBe(true)

    const confirmationInput =
      document.body.querySelector<HTMLInputElement>('#revoke-key-name')
    expect(confirmationInput).not.toBeNull()
    confirmationInput!.value = 'Demo discovery agent'
    confirmationInput!.dispatchEvent(new Event('input', { bubbles: true }))
    await nextTick()
    expect(revokeButton?.disabled).toBe(false)
    wrapper.unmount()
  })

  it('requires the exact asset name before permanent deletion', async () => {
    const wrapper = await mountSuspended(AppDeleteDialog, {
      attachTo: document.body,
      props: {
        modelValue: true,
        assetType: 'machine',
        assetName: 'payments-api-01',
        description: 'This machine will be removed.',
        impact: 'Its address will disappear from the published list.',
        confirmLabel: 'Delete machine',
      },
    })
    const deleteButton = Array.from(
      document.body.querySelectorAll('button'),
    ).find((button) => button.textContent?.includes('Delete machine'))

    expect(document.body.textContent).toContain(
      'Type payments-api-01 to confirm',
    )
    expect(deleteButton?.disabled).toBe(true)

    const confirmationInput = document.body.querySelector<HTMLInputElement>(
      '#delete-asset-confirmation',
    )
    expect(confirmationInput).not.toBeNull()
    confirmationInput!.value = 'payments-api-01'
    confirmationInput!.dispatchEvent(new Event('input', { bubbles: true }))
    await nextTick()
    expect(deleteButton?.disabled).toBe(false)
    wrapper.unmount()
  })

  it('renders every create dialog contract', async () => {
    const machine = await mountSuspended(MachineDialog, {
      props: { modelValue: true, ecosystems: [ecosystem], loading: false },
    })
    const ecosystemDialog = await mountSuspended(EcosystemDialog, {
      props: { modelValue: true, loading: false },
    })
    const key = await mountSuspended(IngestionKeyDialog, {
      props: { modelValue: true, loading: false },
    })
    const trusted = await mountSuspended(TrustedServiceDialog, {
      props: { modelValue: true, loading: false },
    })

    expect(machine.exists()).toBe(true)
    expect(ecosystemDialog.exists()).toBe(true)
    expect(key.exists()).toBe(true)
    expect(trusted.exists()).toBe(true)
  })

  it('renders the trusted catalogue empty state and actions', async () => {
    const wrapper = await mountSuspended(TrustedServicesTable, {
      props: {
        items: [],
        loading: false,
        canEdit: true,
        canResolve: true,
      },
    })
    expect(wrapper.text()).toContain('No trusted services')
    expect(wrapper.text()).toContain('Resolve now')
    expect(wrapper.text()).not.toContain('Cards')
  })

  it('renders trusted services as a table and supports editing', async () => {
    const catalogue = await mountSuspended(TrustedServicesTable, {
      props: {
        items: [trustedService],
        loading: false,
        canEdit: true,
        canResolve: true,
      },
    })
    expect(catalogue.find('table').exists()).toBe(true)
    expect(catalogue.text()).toContain('Core DNS')

    const dialog = await mountSuspended(TrustedServiceDialog, {
      props: {
        modelValue: true,
        loading: false,
        service: trustedService,
      },
    })
    expect(dialog.exists()).toBe(true)
    expect(document.body.textContent).toContain('Edit trusted service')
    expect(document.body.textContent).toContain('Save changes')
  })
})
