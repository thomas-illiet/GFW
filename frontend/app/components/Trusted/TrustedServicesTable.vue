<script setup lang="ts">
import type { TrustedHostname, TrustedService } from '~/types/api'

const props = defineProps<{
  items: TrustedService[]
  canEdit: boolean
  canResolve: boolean
  loading: boolean
  hasServices: boolean
  hasFilters: boolean
}>()
defineEmits<{
  create: []
  edit: [TrustedService]
  remove: [TrustedService]
  showResolved: [TrustedService]
  resolve: []
  export: []
}>()

function currentIPs(host: TrustedHostname) {
  return host.IPs.filter((ip) => ip.IsCurrent).map((ip) => ip.IPv4)
}
function serviceDNSStatus(service: TrustedService) {
  if (service.Hostnames.some((host) => host.Status === 'stale')) return 'stale'
  if (service.Hostnames.some((host) => host.Status === 'pending'))
    return 'pending'
  return 'fresh'
}
function statusColor(status: string) {
  return status === 'fresh'
    ? 'success'
    : status === 'stale'
      ? 'warning'
      : 'info'
}
function resolvedCount(service: TrustedService) {
  return new Set(service.Hostnames.flatMap(currentIPs)).size
}
function latestSuccess(service: TrustedService) {
  return service.Hostnames.map((host) => host.LastSuccessAt)
    .filter((value): value is string => Boolean(value))
    .sort((left, right) => Date.parse(right) - Date.parse(left))[0]
}
function latestError(service: TrustedService) {
  return service.Hostnames.find((host) => host.LastError)?.LastError ?? ''
}
const tableItems = computed(() =>
  props.items.map((service) => ({
    ...service,
    HostnameSort: service.Hostnames.map((host) => host.Hostname).join(' '),
    IPv4Count: resolvedCount(service),
    DNSStatus: serviceDNSStatus(service),
    DNSLastSuccess: latestSuccess(service),
    DNSError: latestError(service),
  })),
)
const headers = computed(() => [
  { title: 'Service', key: 'Name' },
  { title: 'DNS hostnames', key: 'HostnameSort' },
  { title: 'IP addresses', key: 'IPv4Count', align: 'center' as const },
  { title: 'DNS status', key: 'DNSStatus', align: 'center' as const },
  { title: 'Service status', key: 'IsActive', align: 'center' as const },
  {
    title: 'Actions',
    key: 'actions',
    align: 'center' as const,
    sortable: false,
  },
])
</script>

<template>
  <AppSectionCard
    title="Trusted service catalogue"
    subtitle="Active services are included in every ecosystem list."
    icon="mdi-shield-check-outline"
  >
    <template #actions>
      <AppExportButton :disabled="!items.length" @export="$emit('export')" />
      <v-btn
        v-if="canResolve"
        variant="tonal"
        color="primary"
        prepend-icon="mdi-dns-outline"
        :loading="loading"
        :disabled="!hasServices"
        @click="$emit('resolve')"
      >
        Resolve now
      </v-btn>
      <v-btn
        v-if="canEdit"
        color="primary"
        prepend-icon="mdi-plus"
        @click="$emit('create')"
      >
        Add service
      </v-btn>
    </template>

    <v-data-table
      v-if="items.length"
      class="services-table services-table--actions"
      :headers="headers"
      :items="tableItems"
      item-value="ID"
      :items-per-page="10"
    >
      <template #[`item.Name`]="{ item: service }">
        <div class="table-service">
          <strong>{{ service.Name }}</strong>
          <small>{{ service.Description }}</small>
        </div>
      </template>
      <template #[`item.HostnameSort`]="{ item: service }">
        <div class="hostname-list">
          <span v-for="host in service.Hostnames" :key="host.ID">
            <v-icon
              :icon="
                host.Status === 'fresh'
                  ? 'mdi-check-network-outline'
                  : 'mdi-clock-alert-outline'
              "
              :color="statusColor(host.Status)"
              size="small"
            />
            {{ host.Hostname }}
          </span>
        </div>
      </template>
      <template #[`item.IPv4Count`]="{ item: service }">
        <v-chip size="small" variant="tonal">
          {{ service.IPv4Count }}
        </v-chip>
      </template>
      <template #[`item.DNSStatus`]="{ item: service }">
        <div class="dns-health" :title="service.DNSError || undefined">
          <v-chip
            :color="statusColor(service.DNSStatus)"
            size="small"
            variant="tonal"
          >
            {{ service.DNSStatus }}
          </v-chip>
          <AppRelativeTime
            :value="service.DNSLastSuccess"
            prefix="Resolved"
            empty="Never resolved"
          />
        </div>
      </template>
      <template #[`item.IsActive`]="{ item: service }">
        <v-chip
          :color="service.IsActive ? 'success' : 'secondary'"
          size="small"
        >
          {{ service.IsActive ? 'Active' : 'Inactive' }}
        </v-chip>
      </template>
      <template #[`item.actions`]="{ item: service }">
        <AppActionMenu :aria-label="`Actions for ${service.Name}`">
          <v-list-item
            prepend-icon="mdi-ip-network-outline"
            title="View resolved addresses"
            @click="$emit('showResolved', service)"
          />
          <template v-if="canEdit">
            <v-divider class="my-1" />
            <v-list-item
              prepend-icon="mdi-pencil-outline"
              title="Edit service"
              @click="$emit('edit', service)"
            />
            <v-list-item
              class="text-error"
              prepend-icon="mdi-delete-outline"
              title="Delete service"
              @click="$emit('remove', service)"
            />
          </template>
        </AppActionMenu>
      </template>
    </v-data-table>

    <AppEmptyState
      v-else
      icon="mdi-shield-off-outline"
      :title="
        hasFilters ? 'No matching trusted services' : 'No trusted services'
      "
      :text="
        hasFilters
          ? 'Adjust or clear the filters to see more trusted services.'
          : 'Add shared DNS services that are always allowed.'
      "
    />
  </AppSectionCard>
</template>

<style scoped>
.services-table th,
.services-table td {
  vertical-align: middle;
}
.services-table :deep(.v-table__wrapper > table) {
  min-width: 980px;
}
.services-table :deep(th:first-child),
.services-table :deep(td:first-child) {
  min-width: 280px;
}
.services-table :deep(th:nth-child(2)),
.services-table :deep(td:nth-child(2)) {
  min-width: 250px;
}
@media (max-width: 1279px) {
  .services-table--actions :deep(th:last-child),
  .services-table--actions :deep(td:last-child) {
    position: sticky;
    right: 0;
    z-index: 1;
    min-width: 104px;
    background: rgb(var(--gfw-surface));
    box-shadow: -12px 0 18px -18px rgba(var(--gfw-shadow), 0.8);
  }
  .services-table--actions :deep(th:last-child) {
    z-index: 2;
    background: rgb(var(--gfw-muted)) !important;
  }
}
.table-service,
.hostname-list {
  display: grid;
  gap: 4px;
  min-width: 190px;
}
.table-service small {
  max-width: 280px;
  color: rgb(var(--gfw-text-muted));
}
.hostname-list span {
  display: flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
}
.dns-health {
  display: grid;
  justify-items: center;
  gap: 5px;
}
.dns-health :deep(.relative-time) {
  font-size: 0.72rem;
}
@media (max-width: 700px) {
  .services-table :deep(.v-table__wrapper > table) {
    min-width: 620px;
  }
  .services-table :deep(th:nth-child(3)),
  .services-table :deep(td:nth-child(3)),
  .services-table :deep(th:nth-child(4)),
  .services-table :deep(td:nth-child(4)) {
    display: none;
  }
  .services-table :deep(th:first-child),
  .services-table :deep(td:first-child) {
    min-width: 210px;
  }
  .services-table :deep(th:nth-child(2)),
  .services-table :deep(td:nth-child(2)) {
    min-width: 200px;
  }
  .table-service small {
    display: none;
  }
}
</style>
