<script setup lang="ts">
const props = defineProps<{
  loadedAt?: string | null
  loading: boolean
  exportDisabled: boolean
}>()

const emit = defineEmits<{
  refresh: []
  export: []
}>()

const window = defineModel<'24h' | '7d' | '30d'>({ required: true })
</script>

<template>
  <v-card class="dashboard-toolbar">
    <div class="dashboard-toolbar__freshness">
      <v-icon icon="mdi-database-clock-outline" color="primary" size="20" />
      <div>
        <small>Data freshness</small>
        <AppRelativeTime
          :value="props.loadedAt"
          prefix="Updated"
          empty="Loading data"
        />
      </div>
    </div>
    <div class="dashboard-toolbar__actions">
      <DashboardWindowSelect v-model="window" />
      <v-btn
        variant="text"
        prepend-icon="mdi-refresh"
        :loading="props.loading"
        @click="emit('refresh')"
      >
        Refresh
      </v-btn>
      <AppExportButton
        :disabled="props.exportDisabled"
        @export="emit('export')"
      />
    </div>
  </v-card>
</template>

<style scoped>
.dashboard-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 10px 12px 10px 16px;
}
.dashboard-toolbar__freshness,
.dashboard-toolbar__actions {
  display: flex;
  align-items: center;
  gap: 12px;
}
.dashboard-toolbar__freshness small {
  display: block;
  margin-bottom: 2px;
  color: rgb(var(--gfw-text-muted));
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.dashboard-toolbar__actions :deep(.v-btn-toggle) {
  margin-right: 4px;
}
@media (max-width: 800px) {
  .dashboard-toolbar,
  .dashboard-toolbar__actions {
    align-items: stretch;
    flex-direction: column;
  }
  .dashboard-toolbar__actions :deep(.v-btn-toggle) {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
  }
}
</style>
