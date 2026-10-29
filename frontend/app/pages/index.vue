<script setup lang="ts">
import { exportCsv } from '~/utils/export'

definePageMeta({ role: 'viewer' })
const window = shallowRef<'24h' | '7d' | '30d'>('24h')
const dashboard = useDashboard()
const { data, error, loadedAt, loading } = dashboard
watch(window, (value) => void dashboard.load(value).catch(() => undefined), {
  immediate: true,
})
function refreshDashboard() {
  void dashboard.load(window.value).catch(() => undefined)
}
function exportDashboard() {
  if (!data.value) return
  const d = data.value
  exportCsv(
    `great-firewall-dashboard-${window.value}.csv`,
    ['Section', 'Metric', 'Value'],
    [
      ['requests', 'total', d.requests.total],
      ['requests', 'previous total', d.requests.previousTotal],
      ['requests', 'change percent', d.requests.changePercent],
      ['requests', 'success rate', d.requests.successRate],
      ['requests', 'error rate', d.requests.errorRate],
      ['requests', 'not modified', d.requests.notModified],
      ['requests', 'average latency ms', d.requests.averageLatencyMs],
      ['requests', 'p95 latency ms', d.requests.p95LatencyMs],
      ['activity', 'active', d.activity.active],
      ['activity', 'passive', d.activity.passive],
      ['activity', 'total', d.activity.total],
      ['activity', 'active rate', d.activity.activeRate],
      ...d.ecosystems.flatMap((item) => [
        ['ecosystem', `${item.name} requests`, item.requests],
        ['ecosystem', `${item.name} unique IPs`, item.uniqueIps],
      ]),
      ...d.formats.map((item) => ['format', item.name, item.count]),
      ...d.sources.map((item) => ['source', item.name, item.count]),
    ],
  )
}
</script>
<template>
  <v-container fluid class="gfw-page"
    ><v-row
      ><v-col cols="12"
        ><AppPageHeader
          title="Firewall operations"
          copy="See traffic health, address coverage and the ecosystems that need attention."
          icon="mdi-shield-search"
          :stat="data ? `${data.activity.activeRate.toFixed(1)}%` : '—'"
          stat-label="Address coverage" /></v-col
      ><v-col cols="12"
        ><DashboardToolbar
          v-model="window"
          :loaded-at="loadedAt"
          :loading="loading"
          :export-disabled="!data"
          @refresh="refreshDashboard"
          @export="exportDashboard" /></v-col
      ><v-col v-if="error" cols="12"
        ><v-alert type="error" variant="tonal">{{ error }}</v-alert></v-col
      ><template v-if="data"
        ><v-col cols="12"
          ><DashboardRequestKpis
            :value="data.requests"
            :activity="data.activity" /></v-col
        ><v-col cols="12" lg="8"
          ><DashboardRequestChart
            class="fill-height"
            :items="data.series" /></v-col
        ><v-col cols="12" lg="4"
          ><DashboardActivityKpis
            class="fill-height"
            :value="data.activity" /></v-col
        ><v-col cols="12"><DashboardBreakdowns :data="data" /></v-col></template
      ><v-col v-else cols="12"
        ><v-skeleton-loader type="heading, image, table" /></v-col></v-row
  ></v-container>
</template>
