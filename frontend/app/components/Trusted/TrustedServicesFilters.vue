<script setup lang="ts">
type ServiceStatus = '' | 'active' | 'inactive'
type DNSStatus = '' | 'fresh' | 'stale' | 'pending'

const search = defineModel<string>('search', { required: true })
const serviceStatus = defineModel<ServiceStatus>('serviceStatus', {
  required: true,
})
const dnsStatus = defineModel<DNSStatus>('dnsStatus', { required: true })

const serviceStatusItems: Array<{ title: string; value: ServiceStatus }> = [
  { title: 'All service statuses', value: '' },
  { title: 'Active', value: 'active' },
  { title: 'Inactive', value: 'inactive' },
]
const dnsStatusItems: Array<{ title: string; value: DNSStatus }> = [
  { title: 'All DNS statuses', value: '' },
  { title: 'Fresh', value: 'fresh' },
  { title: 'Stale', value: 'stale' },
  { title: 'Pending', value: 'pending' },
]
</script>

<template>
  <v-card class="filters">
    <v-text-field
      v-model="search"
      density="comfortable"
      label="Search trusted services"
      placeholder="Service, hostname, description or IP address"
      prepend-inner-icon="mdi-magnify"
      clearable
      hide-details
    />
    <v-select
      v-model="serviceStatus"
      :items="serviceStatusItems"
      item-title="title"
      item-value="value"
      density="comfortable"
      label="Service status"
      prepend-inner-icon="mdi-shield-check-outline"
      hide-details
    />
    <v-select
      v-model="dnsStatus"
      :items="dnsStatusItems"
      item-title="title"
      item-value="value"
      density="comfortable"
      label="DNS status"
      prepend-inner-icon="mdi-dns-outline"
      hide-details
    />
  </v-card>
</template>

<style scoped>
.filters {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(210px, 1fr) minmax(210px, 1fr);
  align-items: center;
  gap: 12px;
  padding: 12px;
}

@media (max-width: 900px) {
  .filters {
    grid-template-columns: 1fr;
  }
}
</style>
