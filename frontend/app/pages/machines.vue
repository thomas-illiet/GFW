<script setup lang="ts">
import type { Machine, MachineSource } from '~/types/api'
import { exportCsv } from '~/utils/export'
definePageMeta({ role: 'viewer' })
const auth = useAuthStore(),
  r = useResources(),
  notifications = useNotifications()
const { ecosystems, machines, loading, error } = r
const search = shallowRef(''),
  ecosystem = shallowRef(''),
  source = shallowRef<MachineSource>(''),
  dialog = shallowRef(false),
  deleteDialog = shallowRef(false),
  machineToDelete = shallowRef<Machine | null>(null)
const filtered = computed(() =>
  r.machines.value.filter(
    (m) =>
      (!search.value ||
        `${m.Hostname} ${m.IPv4}`
          .toLowerCase()
          .includes(search.value.toLowerCase())) &&
      (!ecosystem.value || m.EcosystemID === ecosystem.value) &&
      (!source.value ||
        (source.value === 'managed' ? m.IsManaged : !m.IsManaged)),
  ),
)
onMounted(
  () =>
    void Promise.all([r.loadEcosystems(), r.loadMachines()]).catch(
      () => undefined,
    ),
)
async function save(p: {
  hostname: string
  ecosystemID: string
  ipv4: string
}) {
  await r.mutation('/api/v1/admin/machines', 'POST', p)
  dialog.value = false
  await r.loadMachines()
  notifications.success('Machine saved')
}
function requestRemove(machine: Machine) {
  machineToDelete.value = machine
  deleteDialog.value = true
}
async function confirmRemove() {
  if (!machineToDelete.value) return
  await r.mutation(
    `/api/v1/admin/machines/${machineToDelete.value.ID}`,
    'DELETE',
  )
  deleteDialog.value = false
  machineToDelete.value = null
  await r.loadMachines()
  notifications.success('Machine deleted')
}
function exportMachines() {
  exportCsv(
    'great-firewall-machines.csv',
    ['Hostname', 'IP address', 'Ecosystem', 'Source', 'Last used', 'Last sync'],
    filtered.value.map((machine) => [
      machine.Hostname,
      machine.IPv4,
      machine.EcosystemName,
      machine.IsManaged ? 'managed' : 'manual',
      machine.LastUsedAt,
      machine.LastSyncAt,
    ]),
  )
}
</script>
<template>
  <v-container fluid class="gfw-page"
    ><v-row
      ><v-col cols="12"
        ><AppPageHeader
          title="Machines"
          copy="Manage the IP address inventory grouped by product ecosystem."
          icon="mdi-server-network"
          :stat="String(machines.length)"
          stat-label="Addresses" /></v-col
      ><v-col cols="12"
        ><MachinesMachineFilters
          v-model:search="search"
          v-model:ecosystem="ecosystem"
          v-model:source="source"
          :ecosystems="ecosystems" /></v-col
      ><v-col v-if="error" cols="12"
        ><v-alert type="error">{{ error }}</v-alert></v-col
      ><v-col cols="12"
        ><MachinesTable
          :items="filtered"
          :loading="loading"
          :can-edit="auth.can('editor')"
          @create="dialog = true"
          @export="exportMachines"
          @remove="requestRemove" /></v-col></v-row
    ><MachinesMachineDialog
      v-model="dialog"
      :ecosystems="ecosystems"
      :loading="loading"
      @save="save" /><AppDeleteDialog
      v-model="deleteDialog"
      asset-type="machine"
      :asset-name="machineToDelete?.Hostname"
      description="This machine and its address will be removed from the inventory."
      impact="The address will immediately disappear from its ecosystem firewall list."
      confirm-label="Delete machine"
      :loading="loading"
      @confirm="confirmRemove"
  /></v-container>
</template>
