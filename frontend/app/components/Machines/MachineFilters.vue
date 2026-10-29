<script setup lang="ts">
import type { Ecosystem, MachineSource } from '~/types/api'
defineProps<{ ecosystems: Ecosystem[] }>()
const search = defineModel<string>('search', { required: true })
const ecosystem = defineModel<string>('ecosystem', { required: true })
const source = defineModel<MachineSource>('source', { required: true })
const sourceItems: Array<{ title: string; value: MachineSource }> = [
  { title: 'All sources', value: '' },
  { title: 'Managed', value: 'managed' },
  { title: 'Manual', value: 'manual' },
]
</script>
<template>
  <v-card class="filters"
    ><v-text-field
      v-model="search"
      density="comfortable"
      label="Search hostname or IP address"
      prepend-inner-icon="mdi-magnify"
      hide-details
      clearable /><v-select
      v-model="ecosystem"
      density="comfortable"
      :items="ecosystems"
      item-title="DisplayName"
      item-value="ID"
      label="Ecosystem"
      hide-details
      clearable /><v-select
      v-model="source"
      density="comfortable"
      :items="sourceItems"
      label="Source"
      prepend-inner-icon="mdi-source-branch"
      hide-details
  /></v-card>
</template>
<style scoped>
.filters {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr;
  gap: 12px;
  padding: 12px;
}
@media (max-width: 700px) {
  .filters {
    grid-template-columns: 1fr;
  }
}
</style>
