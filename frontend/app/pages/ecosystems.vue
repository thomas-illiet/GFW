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
  deleteDialog = shallowRef(false),
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
function requestDelete(ecosystem: Ecosystem) {
  selectedEcosystem.value = ecosystem
  deleteDialog.value = true
}
async function confirmDelete() {
  const ecosystem = selectedEcosystem.value
  if (!ecosystem) return
  await r.mutation(`/api/v1/admin/ecosystems/${ecosystem.ID}`, 'DELETE')
  deleteDialog.value = false
  selectedEcosystem.value = null
  await r.loadEcosystems()
  notifications.success('Ecosystem deleted')
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
            :items-per-page="10"
            ><template #[`item.MachineCount`]="{ item: e }">
              <v-btn
                prepend-icon="mdi-server-network"
                size="small"
                variant="tonal"
                color="primary"
                :to="{ path: '/machines', query: { ecosystem: e.ID } }"
                :aria-label="`View ${e.MachineCount} machines in ${e.DisplayName}`"
                >{{ e.MachineCount }}</v-btn
              >
            </template>
            <template #[`item.IsActive`]="{ item: e }">
              <v-btn
                :color="e.IsActive ? 'success' : 'secondary'"
                size="small"
                variant="tonal"
                :disabled="!auth.can('admin')"
                :aria-label="`${e.IsActive ? 'Disable' : 'Enable'} ${e.DisplayName}`"
                @click="requestStatusChange(e)"
                >{{ e.IsActive ? 'Active' : 'Inactive' }}</v-btn
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
              <AppActionMenu :aria-label="`Actions for ${e.DisplayName}`">
                <v-list-item
                  prepend-icon="mdi-server-network"
                  title="View machines"
                  :to="{ path: '/machines', query: { ecosystem: e.ID } }"
                />
                <template v-if="auth.can('admin')">
                  <v-list-item
                    :prepend-icon="
                      e.IsActive
                        ? 'mdi-pause-circle-outline'
                        : 'mdi-play-circle-outline'
                    "
                    :title="
                      e.IsActive ? 'Disable ecosystem' : 'Enable ecosystem'
                    "
                    @click="requestStatusChange(e)"
                  />
                  <v-divider class="my-1" />
                  <v-list-item
                    class="text-error"
                    prepend-icon="mdi-delete-outline"
                    title="Delete ecosystem"
                    :subtitle="
                      e.MachineCount
                        ? `Also delete ${e.MachineCount} assigned machine${e.MachineCount === 1 ? '' : 's'}`
                        : 'Permanently remove this ecosystem'
                    "
                    @click="requestDelete(e)"
                  />
                </template>
              </AppActionMenu>
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
      @save="create" /><AppDeleteDialog
      v-if="selectedEcosystem?.IsActive"
      v-model="statusDialog"
      asset-type="ecosystem"
      :asset-name="selectedEcosystem?.DisplayName"
      dialog-title="Disable ecosystem?"
      eyebrow="Firewall list interruption"
      description="This ecosystem will stop publishing its public firewall list."
      warning-title="Firewall consumers will lose access"
      impact="The list remains unavailable until an administrator enables the ecosystem again."
      confirm-label="Disable ecosystem"
      confirm-color="warning"
      confirm-icon="mdi-pause-circle-outline"
      icon="mdi-pause-circle-outline"
      :loading="loading"
      @confirm="confirmStatusChange" /><AppConfirmDialog
      v-else
      v-model="statusDialog"
      title="Enable ecosystem?"
      text="Its public firewall list will become available again immediately."
      :subject="selectedEcosystem?.DisplayName"
      confirm-label="Enable"
      confirm-icon="mdi-play-circle-outline"
      confirm-color="success"
      icon="mdi-play-circle-outline"
      :loading="loading"
      @confirm="confirmStatusChange" /><AppDeleteDialog
      v-model="deleteDialog"
      asset-type="ecosystem"
      :asset-name="selectedEcosystem?.DisplayName"
      description="This ecosystem and its firewall list configuration will be permanently removed."
      :impact="
        selectedEcosystem?.MachineCount
          ? `${selectedEcosystem.MachineCount} assigned machine${selectedEcosystem.MachineCount === 1 ? '' : 's'} will also be permanently deleted.`
          : 'The ecosystem will be permanently removed. This action cannot be undone.'
      "
      confirm-label="Delete ecosystem"
      :loading="loading"
      @confirm="confirmDelete"
  /></v-container>
</template>

<style scoped>
.status-chip--clickable {
  cursor: pointer;
}
</style>
