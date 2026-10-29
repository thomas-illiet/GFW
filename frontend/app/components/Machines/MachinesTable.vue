<script setup lang="ts">
import type { Machine } from '~/types/api'
defineProps<{ items: Machine[]; loading: boolean; canEdit: boolean }>()
defineEmits<{ create: []; remove: [Machine]; export: [] }>()
const headers = [
  { title: 'Machine', key: 'Hostname' },
  { title: 'IP address', key: 'IPv4' },
  { title: 'Source', key: 'IsManaged', align: 'center' as const },
  { title: 'Last used', key: 'LastUsedAt', align: 'center' as const },
  { title: 'Last sync', key: 'LastSyncAt', align: 'center' as const },
  {
    title: 'Actions',
    key: 'actions',
    sortable: false,
    align: 'center' as const,
  },
]
</script>
<template>
  <AppSectionCard
    title="Machine addresses"
    :subtitle="`${items.length} addresses configured`"
    icon="mdi-server-network"
    ><template #actions
      ><AppExportButton
        :disabled="!items.length"
        @export="$emit('export')"
      /><v-btn
        v-if="canEdit"
        color="primary"
        prepend-icon="mdi-plus"
        @click="$emit('create')"
        >Add machine</v-btn
      ></template
    ><v-data-table
      class="action-table"
      :headers="headers"
      :items="items"
      :loading="loading"
      item-value="ID"
      ><template #[`item.Hostname`]="{ item }"
        ><div class="machine-identity">
          <strong>{{ item.Hostname }}</strong>
          <small>{{ item.EcosystemName }}</small>
        </div></template
      ><template #[`item.IPv4`]="{ item }"
        ><code>{{ item.IPv4 }}</code></template
      ><template #[`item.IsManaged`]="{ item }"
        ><v-chip
          :color="item.IsManaged ? 'info' : 'secondary'"
          size="small"
          variant="tonal"
          >{{ item.IsManaged ? 'managed' : 'manual' }}</v-chip
        ></template
      ><template #[`item.LastUsedAt`]="{ item }"
        ><AppRelativeTime
          :value="item.LastUsedAt"
          prefix="Used"
          empty="Never used" /></template
      ><template #[`item.LastSyncAt`]="{ item }"
        ><AppRelativeTime
          :value="item.LastSyncAt"
          prefix="Synced"
          empty="Never synced" /></template
      ><template #[`item.actions`]="{ item }"
        ><v-btn
          v-if="canEdit && !item.IsManaged"
          icon="mdi-delete-outline"
          color="error"
          variant="text"
          aria-label="Delete machine"
          @click="$emit('remove', item)" /></template
      ><template #no-data
        ><AppEmptyState
          icon="mdi-server-network-off"
          title="No machines found"
          text="Add a machine or adjust the current filters." /></template></v-data-table
  ></AppSectionCard>
</template>

<style scoped>
.machine-identity {
  display: grid;
  gap: 2px;
  min-width: 170px;
}
.machine-identity small {
  color: rgb(var(--gfw-text-muted));
  font-size: 0.75rem;
}
</style>
