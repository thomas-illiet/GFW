import type {
  Dashboard,
  Ecosystem,
  EffectiveIP,
  IngestionKey,
  ListResponse,
  Machine,
  TrustedService,
} from '~/types/api'
export function useResources() {
  const api = useApi()
  const ecosystems = shallowRef<Ecosystem[]>([]),
    machines = shallowRef<Machine[]>([]),
    trusted = shallowRef<TrustedService[]>([]),
    keys = shallowRef<IngestionKey[]>([])
  async function loadEcosystems() {
    const r = await api.request<ListResponse<Ecosystem>>(
      '/api/v1/admin/ecosystems',
    )
    ecosystems.value = r.items ?? []
  }
  async function loadMachines(params = '') {
    const r = await api.request<ListResponse<Machine>>(
      `/api/v1/admin/machines${params}`,
    )
    machines.value = r.items ?? []
  }
  async function loadTrusted() {
    const r = await api.request<ListResponse<TrustedService>>(
      '/api/v1/admin/trusted-services',
    )
    trusted.value = (r.items ?? []).map((service) => ({
      ...service,
      Hostnames: (service.Hostnames ?? []).map((hostname) => ({
        ...hostname,
        IPs: hostname.IPs ?? [],
      })),
    }))
  }
  async function loadKeys() {
    const r = await api.request<ListResponse<IngestionKey>>(
      '/api/v1/admin/ingestion-keys',
    )
    keys.value = r.items ?? []
  }
  async function mutation<T>(
    path: string,
    method: 'POST' | 'PATCH' | 'DELETE',
    body?: Record<string, unknown>,
  ) {
    return await api.request<T>(path, { method, body })
  }
  async function effective(ecosystem: string) {
    return await api.request<{
      ecosystem: string
      version: string
      generatedAt: string
      items: EffectiveIP[]
    }>(
      `/api/v1/firewall/ecosystems/${encodeURIComponent(ecosystem)}/ips?format=json`,
    )
  }
  return {
    ...api,
    ecosystems,
    machines,
    trusted,
    keys,
    loadEcosystems,
    loadMachines,
    loadTrusted,
    loadKeys,
    mutation,
    effective,
  }
}
export function useDashboard() {
  const api = useApi()
  const data = shallowRef<Dashboard | null>(null)
  const loadedAt = shallowRef<string | null>(null)
  async function load(window: '24h' | '7d' | '30d') {
    data.value = await api.request<Dashboard>(
      `/api/v1/admin/dashboard?window=${window}`,
    )
    loadedAt.value = new Date().toISOString()
  }
  return { ...api, data: readonly(data), loadedAt: readonly(loadedAt), load }
}
