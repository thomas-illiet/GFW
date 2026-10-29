<script setup lang="ts">
import type { EffectiveIP } from '~/types/api'
import { exportCsv } from '~/utils/export'
definePageMeta({ role: 'viewer' })
const route = useRoute(),
  r = useResources(),
  ecosystem = shallowRef(''),
  items = shallowRef<EffectiveIP[]>([]),
  version = shallowRef('')
let requestSequence = 0
const { ecosystems, loading, error } = r
const headers = [
  { title: 'IP address', key: 'IPv4', align: 'center' as const },
  { title: 'Source', key: 'SourceType', align: 'center' as const },
  { title: 'Hostname', key: 'Hostname' },
  { title: 'Service', key: 'Service' },
  { title: 'Health', key: 'Stale', align: 'center' as const },
]
onMounted(() => {
  void loadInitialEcosystem().catch(() => undefined)
})
watch(ecosystem, (selected) => {
  void loadEffectiveList(selected).catch(() => undefined)
})
async function loadInitialEcosystem() {
  await r.loadEcosystems()
  const requested =
    typeof route.query.ecosystem === 'string' ? route.query.ecosystem : ''
  if (!requested) return
  const selected = ecosystems.value.find(
    (item) => item.Name === requested && item.IsActive,
  )
  if (!selected) return
  ecosystem.value = selected.Name
}
async function loadEffectiveList(selected: string) {
  const request = ++requestSequence
  if (!selected) {
    items.value = []
    version.value = ''
    return
  }
  items.value = []
  version.value = ''
  const result = await r.effective(selected)
  if (request !== requestSequence) return
  items.value = result.items
  version.value = result.version
}
function exportEffectiveList() {
  exportCsv(
    `great-firewall-${ecosystem.value || 'effective'}-ips.csv`,
    [
      'IP address',
      'Source',
      'Hostname',
      'Service',
      'Health',
      'Last used',
      'Last sync',
    ],
    items.value.map((item) => [
      item.IPv4,
      item.SourceType,
      item.Hostname,
      item.Service,
      item.Stale ? 'stale' : 'fresh',
      item.LastUsedAt,
      item.LastSyncAt,
    ]),
  )
}
</script>
<template>
  <v-container fluid class="gfw-page"
    ><v-row
      ><v-col cols="12"
        ><AppPageHeader
          title="Effective lists"
          copy="View exactly what a firewall receives for a product ecosystem."
          icon="mdi-format-list-checks"
          :stat="String(items.length)"
          stat-label="Entries" /></v-col
      ><v-col v-if="error" cols="12"
        ><v-alert type="error" variant="tonal">{{ error }}</v-alert></v-col
      ><v-col cols="12"
        ><v-card class="effective-controls pa-3"
          ><v-select
            v-model="ecosystem"
            class="effective-controls__select"
            density="comfortable"
            :items="ecosystems.filter((e) => e.IsActive)"
            item-title="DisplayName"
            item-value="Name"
            label="Ecosystem"
            :loading="loading"
            hide-details /><AppExportButton
            class="effective-controls__action"
            :disabled="!items.length"
            @export="exportEffectiveList" /></v-card></v-col
      ><v-col cols="12"
        ><AppSectionCard
          title="Published addresses"
          :subtitle="
            version
              ? `Version ${version.slice(0, 12)}`
              : 'Select an ecosystem to load its list.'
          "
          icon="mdi-ip-network-outline"
          ><v-progress-linear
            v-if="loading && ecosystem"
            color="primary"
            indeterminate
            rounded
            class="mb-4" /><v-data-table
            v-if="items.length"
            :headers="headers"
            :items="items"
            item-value="SourceID"
            :items-per-page="-1"
            hide-default-footer
            ><template #[`item.IPv4`]="{ item: i }">
              <code>{{ i.IPv4 }}</code>
            </template>
            <template #[`item.SourceType`]="{ item: i }">
              <v-chip size="small" color="primary" variant="tonal">{{
                i.SourceType
              }}</v-chip>
            </template>
            <template #[`item.Service`]="{ item: i }">{{
              i.Service || '—'
            }}</template>
            <template #[`item.Stale`]="{ item: i }">
              <v-chip size="small" :color="i.Stale ? 'warning' : 'success'">{{
                i.Stale ? 'stale' : 'fresh'
              }}</v-chip>
            </template></v-data-table
          ><AppEmptyState
            v-else-if="!loading"
            icon="mdi-format-list-checks"
            title="No list loaded"
            text="Choose an ecosystem to load its effective list." /></AppSectionCard></v-col></v-row
  ></v-container>
</template>

<style scoped>
.effective-controls {
  display: flex;
  align-items: center;
  gap: 14px;
}
.effective-controls__select {
  min-width: 260px;
  flex: 1 1 360px;
}
.effective-controls__action {
  min-width: 190px;
  min-height: 48px;
  padding-inline: 20px;
}
@media (max-width: 760px) {
  .effective-controls {
    align-items: stretch;
    flex-direction: column;
  }
  .effective-controls__select {
    min-width: 0;
    flex-basis: auto;
  }
  .effective-controls__action {
    width: 100%;
  }
}
</style>
