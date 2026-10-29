<script setup lang="ts">
import type { DeepReadonly } from 'vue'

import type { Dashboard } from '~/types/api'
import { number, percent } from '~/utils/format'

const props = defineProps<{
  items: DeepReadonly<Dashboard['ecosystems']>
}>()

const total = computed(() =>
  props.items.reduce((sum, item) => sum + item.requests, 0),
)
const tableItems = computed(() =>
  props.items.map((item) => ({
    ...item,
    trafficShare: share(item.requests),
  })),
)
const headers = [
  { title: 'Ecosystem', key: 'name' },
  { title: 'Requests', key: 'requests', align: 'center' as const },
  { title: 'Traffic share', key: 'trafficShare', align: 'center' as const },
  { title: 'Unique IPs', key: 'uniqueIps', align: 'center' as const },
]

function share(requests: number) {
  return total.value ? (requests * 100) / total.value : 0
}

function displayName(name: string) {
  return name
    .split('-')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}
</script>

<template>
  <AppSectionCard
    title="Ecosystem traffic"
    subtitle="Requests and distinct addresses served"
    icon="mdi-cube-outline"
  >
    <v-data-table
      v-if="items.length"
      :headers="headers"
      :items="tableItems"
      item-value="name"
      :items-per-page="5"
      density="comfortable"
    >
      <template #[`item.name`]="{ item }">
        <strong>{{ displayName(item.name) }}</strong>
        <small>{{ item.name }}</small>
      </template>
      <template #[`item.requests`]="{ item }">{{
        number(item.requests)
      }}</template>
      <template #[`item.trafficShare`]="{ item }">
        <div class="traffic-share">
          <v-progress-linear
            :model-value="item.trafficShare"
            color="primary"
            height="7"
            rounded
          />
          <span>{{ percent(item.trafficShare) }}</span>
        </div>
      </template>
      <template #[`item.uniqueIps`]="{ item }">{{
        number(item.uniqueIps)
      }}</template>
    </v-data-table>
    <AppEmptyState
      v-else
      icon="mdi-cube-off-outline"
      title="No ecosystem traffic"
      text="Request activity will appear here once a list is retrieved."
    />
  </AppSectionCard>
</template>

<style scoped>
td strong,
td small {
  display: block;
}
td small {
  margin-top: 2px;
  color: rgb(var(--gfw-text-muted));
  font-size: 0.7rem;
}
.traffic-share {
  display: grid;
  grid-template-columns: minmax(80px, 1fr) 48px;
  align-items: center;
  gap: 8px;
  min-width: 150px;
}
.traffic-share span {
  color: rgb(var(--gfw-text-secondary));
  font-size: 0.75rem;
  text-align: right;
}
</style>
