<script setup lang="ts">
type EcosystemStatus = '' | 'active' | 'inactive'

const search = defineModel<string>('search', { required: true })
const status = defineModel<EcosystemStatus>('status', { required: true })

const statusItems: Array<{ title: string; value: EcosystemStatus }> = [
  { title: 'All statuses', value: '' },
  { title: 'Active', value: 'active' },
  { title: 'Inactive', value: 'inactive' },
]
</script>

<template>
  <v-card class="filters">
    <v-text-field
      v-model="search"
      density="comfortable"
      label="Search ecosystems"
      placeholder="Search by name or identifier"
      prepend-inner-icon="mdi-magnify"
      clearable
      hide-details
    />
    <v-select
      v-model="status"
      :items="statusItems"
      item-title="title"
      item-value="value"
      density="comfortable"
      label="Status"
      prepend-inner-icon="mdi-filter-variant"
      hide-details
    />
  </v-card>
</template>

<style scoped>
.filters {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(220px, 1fr);
  align-items: center;
  gap: 12px;
  padding: 12px;
}

@media (max-width: 700px) {
  .filters {
    grid-template-columns: 1fr;
  }
}
</style>
