<script setup lang="ts">
import type { Ecosystem } from '~/types/api'
import { exportCsv } from '~/utils/export'
type EcosystemStatus = '' | 'active' | 'inactive'
definePageMeta({ role: 'viewer' })
const auth = useAuthStore(),
  r = useResources(),
  notifications = useNotifications(),
  search = shallowRef(''),
  status = shallowRef<EcosystemStatus>(''),
  dialog = shallowRef(false),
  statusDialog = shallowRef(false),
  selectedEcosystem = shallowRef<Ecosystem | null>(null)
const { ecosystems, loading, error } = r
const filteredEcosystems = computed(() => {
  const query = search.value.trim().toLowerCase()
  return ecosystems.value.filter(
    (ecosystem) =>
      (!query ||
        `${ecosystem.DisplayName} ${ecosystem.Name}`
          .toLowerCase()
          .includes(query)) &&
      (!status.value ||
        (status.value === 'active' ? ecosystem.IsActive : !ecosystem.IsActive)),
  )
})
const headers = [
  { title: 'Name', key: 'DisplayName' },
  { title: 'Machines', key: 'MachineCount', align: 'center' as const },
  { title: 'Status', key: 'IsActive', align: 'center' as const },
  { title: 'Last used', key: 'LastUsedAt', align: 'center' as const },
  { title: 'Created at', key: 'CreatedAt', align: 'center' as const },
  {
    title: 'Actions',
    key: 'actions',
    align: 'center' as const,
    sortable: false,
  },
]
onMounted(() => void r.loadEcosystems().catch(() => undefined))
async function create(p: { name: string; displayName: string }) {
  await r.mutation('/api/v1/admin/ecosystems', 'POST', p)
  dialog.value = false
  await r.loadEcosystems()
  notifications.success('Ecosystem created')
}
function requestStatusChange(ecosystem: Ecosystem) {
  if (!auth.can('admin')) return
  selectedEcosystem.value = ecosystem
  statusDialog.value = true
}
async function confirmStatusChange() {
  const e = selectedEcosystem.value
  if (!e) return
  await r.mutation(`/api/v1/admin/ecosystems/${e.ID}`, 'PATCH', {
    displayName: e.DisplayName,
    isActive: !e.IsActive,
  })
  statusDialog.value = false
  selectedEcosystem.value = null
  await r.loadEcosystems()
  notifications.success(`Ecosystem ${e.IsActive ? 'disabled' : 'enabled'}`)
}
function exportEcosystems() {
  exportCsv(
    'great-firewall-ecosystems.csv',
    [
      'Name',
      'Identifier',
      'Machines',
      'Status',
      'Last used',
      'Created at',
      'Updated at',
    ],
    filteredEcosystems.value.map((ecosystem) => [
      ecosystem.DisplayName,
      ecosystem.Name,
      ecosystem.MachineCount,
      ecosystem.IsActive ? 'active' : 'inactive',
      ecosystem.LastUsedAt ?? '',
      ecosystem.CreatedAt,
      ecosystem.UpdatedAt,
    ]),
  )
}
function viewAddresses(ecosystem: Ecosystem) {
  void navigateTo({
    path: '/effective-lists',
    query: { ecosystem: ecosystem.Name },
  })
}
</script>
<template>
  <v-container fluid class="gfw-page"
    ><v-row
      ><v-col cols="12"
        ><AppPageHeader
          title="Ecosystems"
          copy="Define the product boundaries used to publish firewall lists."
          icon="mdi-cube-outline"
          :stat="String(ecosystems.length)"
          stat-label="Ecosystems" /></v-col
      ><v-col cols="12"
        ><EcosystemsEcosystemFilters
          v-model:search="search"
          v-model:status="status" /></v-col
      ><v-col cols="12"
        ><AppSectionCard
          title="Product ecosystems"
          subtitle="Each machine belongs to one ecosystem."
          icon="mdi-cube-outline"
          ><template #actions
            ><AppExportButton
              :disabled="!filteredEcosystems.length"
              @export="exportEcosystems"
            /><v-btn
              v-if="auth.can('admin')"
              color="primary"
              prepend-icon="mdi-plus"
              @click="dialog = true"
              >Create ecosystem</v-btn
            ></template
          ><v-alert v-if="error" type="error" variant="tonal" class="mb-4">{{
            error
          }}</v-alert
          ><v-data-table
            class="action-table"
            :headers="headers"
            :items="filteredEcosystems"
            item-value="ID"
            :items-per-page="-1"
            hide-default-footer
            ><template #[`item.MachineCount`]="{ item: e }">
              <v-chip
                prepend-icon="mdi-server-network"
                size="small"
                variant="tonal"
                >{{ e.MachineCount }}</v-chip
              >
            </template>
            <template #[`item.IsActive`]="{ item: e }">
              <v-chip
                :color="e.IsActive ? 'success' : 'secondary'"
                size="small"
                variant="tonal"
                :class="{ 'status-chip--clickable': auth.can('admin') }"
                :aria-label="`${e.IsActive ? 'Disable' : 'Enable'} ${e.DisplayName}`"
                @click="requestStatusChange(e)"
                >{{ e.IsActive ? 'Active' : 'Inactive' }}</v-chip
              >
            </template>
            <template #[`item.LastUsedAt`]="{ item: e }">
              <AppRelativeTime
                :value="e.LastUsedAt"
                prefix="Used"
                empty="Never used"
              />
            </template>
            <template #[`item.CreatedAt`]="{ item: e }">
              <AppRelativeTime :value="e.CreatedAt" prefix="Created" />
            </template>
            <template #[`item.actions`]="{ item: e }">
              <v-tooltip text="View published addresses" location="top">
                <template #activator="{ props: tooltipProps }">
                  <v-btn
                    v-bind="tooltipProps"
                    icon="mdi-ip-network-outline"
                    size="small"
                    variant="text"
                    color="primary"
                    :aria-label="`View addresses for ${e.DisplayName}`"
                    @click="viewAddresses(e)"
                  />
                </template>
              </v-tooltip>
            </template>
            <template #no-data>
              <AppEmptyState
                icon="mdi-cube-off-outline"
                :title="
                  search || status ? 'No matching ecosystems' : 'No ecosystems'
                "
                :text="
                  search || status
                    ? 'Adjust or clear the filters to see more ecosystems.'
                    : 'Create the first product boundary to publish firewall lists.'
                "
              /> </template></v-data-table></AppSectionCard></v-col></v-row
    ><EcosystemsEcosystemDialog
      v-model="dialog"
      :loading="loading"
      @save="create" /><AppConfirmDialog
      v-model="statusDialog"
      :title="
        selectedEcosystem?.IsActive ? 'Disable ecosystem?' : 'Enable ecosystem?'
      "
      :text="
        selectedEcosystem?.IsActive
          ? 'Its public firewall list will no longer be available until the ecosystem is enabled again.'
          : 'Its public firewall list will become available again immediately.'
      "
      :subject="selectedEcosystem?.DisplayName"
      :confirm-label="selectedEcosystem?.IsActive ? 'Disable' : 'Enable'"
      :confirm-icon="
        selectedEcosystem?.IsActive
          ? 'mdi-pause-circle-outline'
          : 'mdi-play-circle-outline'
      "
      :confirm-color="selectedEcosystem?.IsActive ? 'warning' : 'success'"
      :icon="
        selectedEcosystem?.IsActive
          ? 'mdi-pause-circle-outline'
          : 'mdi-play-circle-outline'
      "
      :loading="loading"
      @confirm="confirmStatusChange"
  /></v-container>
</template>

<style scoped>
.status-chip--clickable {
  cursor: pointer;
}
</style>
