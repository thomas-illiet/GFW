<script setup lang="ts">
import type { TrustedService } from '~/types/api'
import { exportCsv } from '~/utils/export'

type ServiceStatus = '' | 'active' | 'inactive'
type DNSStatus = '' | 'fresh' | 'stale' | 'pending'

definePageMeta({ role: 'viewer' })
const auth = useAuthStore(),
  r = useResources(),
  notifications = useNotifications(),
  search = shallowRef(''),
  serviceStatus = shallowRef<ServiceStatus>(''),
  dnsStatus = shallowRef<DNSStatus>(''),
  dialog = shallowRef(false),
  deleteDialog = shallowRef(false),
  selectedService = shallowRef<TrustedService | null>(null)
const { trusted, loading, error } = r
const hasFilters = computed(() =>
  Boolean(search.value || serviceStatus.value || dnsStatus.value),
)
function serviceDNSStatus(service: TrustedService): Exclude<DNSStatus, ''> {
  if (service.Hostnames.some((hostname) => hostname.Status === 'stale'))
    return 'stale'
  if (service.Hostnames.some((hostname) => hostname.Status === 'pending'))
    return 'pending'
  return 'fresh'
}
const filteredTrusted = computed(() => {
  const query = search.value.trim().toLowerCase()
  return trusted.value.filter((service) => {
    const searchable = [
      service.Name,
      service.Description,
      ...service.Hostnames.flatMap((hostname) => [
        hostname.Hostname,
        ...hostname.IPs.filter((ip) => ip.IsCurrent).map((ip) => ip.IPv4),
      ]),
    ]
      .join(' ')
      .toLowerCase()
    const matchesServiceStatus =
      !serviceStatus.value ||
      (serviceStatus.value === 'active' ? service.IsActive : !service.IsActive)
    const matchesDNSStatus =
      !dnsStatus.value || serviceDNSStatus(service) === dnsStatus.value
    return (
      (!query || searchable.includes(query)) &&
      matchesServiceStatus &&
      matchesDNSStatus
    )
  })
})
onMounted(() => void r.loadTrusted().catch(() => undefined))
async function save(p: {
  name: string
  description: string
  isActive: boolean
  hostnames: string[]
}) {
  const editing = selectedService.value
  await r.mutation(
    editing
      ? `/api/v1/admin/trusted-services/${editing.ID}`
      : '/api/v1/admin/trusted-services',
    editing ? 'PATCH' : 'POST',
    p,
  )
  dialog.value = false
  selectedService.value = null
  await r.loadTrusted()
  notifications.success(
    editing ? 'Trusted service updated' : 'Trusted service created',
  )
}
function createService() {
  selectedService.value = null
  dialog.value = true
}
function editService(service: TrustedService) {
  selectedService.value = service
  dialog.value = true
}
function requestDelete(service: TrustedService) {
  selectedService.value = service
  deleteDialog.value = true
}
async function confirmDelete() {
  if (!selectedService.value) return
  await r.mutation(
    `/api/v1/admin/trusted-services/${selectedService.value.ID}`,
    'DELETE',
  )
  deleteDialog.value = false
  selectedService.value = null
  await r.loadTrusted()
  notifications.success('Trusted service deleted')
}
async function resolve() {
  await r.mutation('/api/v1/admin/trusted-services/resolve', 'POST')
  await r.loadTrusted()
  notifications.success('DNS records refreshed')
}
function exportTrusted() {
  exportCsv(
    'great-firewall-trusted-services.csv',
    [
      'Service',
      'Status',
      'Hostname',
      'DNS status',
      'IP address',
      'Last success',
    ],
    filteredTrusted.value.flatMap((service) =>
      service.Hostnames.flatMap((hostname) => {
        const ips = hostname.IPs.filter((ip) => ip.IsCurrent)
        const values = ips.length ? ips.map((ip) => ip.IPv4) : ['']
        return values.map((ip) => [
          service.Name,
          service.IsActive ? 'active' : 'inactive',
          hostname.Hostname,
          hostname.Status,
          ip,
          hostname.LastSuccessAt,
        ])
      }),
    ),
  )
}
</script>
<template>
  <v-container fluid class="gfw-page"
    ><v-row
      ><v-col cols="12"
        ><AppPageHeader
          title="Trusted services"
          copy="Resolve common service hostnames and include their IP addresses everywhere."
          icon="mdi-shield-check-outline"
          :stat="String(trusted.length)"
          stat-label="Services" /></v-col
      ><v-col v-if="error" cols="12"
        ><v-alert type="error" variant="tonal">{{ error }}</v-alert></v-col
      ><v-col cols="12"
        ><TrustedServicesFilters
          v-model:search="search"
          v-model:service-status="serviceStatus"
          v-model:dns-status="dnsStatus" /></v-col
      ><v-col cols="12"
        ><TrustedServicesTable
          :items="filteredTrusted"
          :has-services="Boolean(trusted.length)"
          :has-filters="hasFilters"
          :loading="loading"
          :can-edit="auth.can('editor')"
          :can-resolve="auth.can('admin')"
          @export="exportTrusted"
          @create="createService"
          @edit="editService"
          @remove="requestDelete"
          @resolve="resolve" /></v-col></v-row
    ><TrustedServiceDialog
      v-model="dialog"
      :loading="loading"
      :service="selectedService"
      @save="save" /><AppDeleteDialog
      v-model="deleteDialog"
      asset-type="trusted service"
      :asset-name="selectedService?.Name"
      description="This trusted service and its DNS resolution history will be removed."
      impact="Its resolved addresses will immediately disappear from every effective firewall list."
      confirm-label="Delete service"
      :loading="loading"
      @confirm="confirmDelete"
  /></v-container>
</template>
