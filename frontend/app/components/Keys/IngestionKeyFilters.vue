<script setup lang="ts">
const search = defineModel<string>('search', { required: true })
const showRevoked = defineModel<boolean>('showRevoked', { required: true })
const statusOptions = [
  { title: 'Active keys', value: false },
  { title: 'All keys (including revoked)', value: true },
]
</script>

<template>
  <v-card class="filters">
    <v-text-field
      v-model="search"
      density="comfortable"
      label="Search name or prefix"
      prepend-inner-icon="mdi-magnify"
      hide-details
      clearable
    />
    <v-select
      v-model="showRevoked"
      :items="statusOptions"
      item-title="title"
      item-value="value"
      density="comfortable"
      label="Key status"
      prepend-inner-icon="mdi-filter-variant"
      hide-details
      variant="outlined"
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
