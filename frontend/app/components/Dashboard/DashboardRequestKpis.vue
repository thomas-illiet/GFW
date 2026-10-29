<script setup lang="ts">
import type { Dashboard } from '~/types/api'
import { number, percent } from '~/utils/format'

const props = defineProps<{
  value: Dashboard['requests']
  activity: Dashboard['activity']
}>()

const trend = computed(() => {
  if (!props.value.previousTotal) return 'No previous traffic'
  const sign = props.value.changePercent > 0 ? '+' : ''
  return `${sign}${props.value.changePercent.toFixed(1)}% vs previous period`
})

const primary = computed(() => [
  {
    label: 'Requests',
    value: number(props.value.total),
    detail: trend.value,
    status: 'Volume change',
    icon: 'mdi-swap-horizontal',
    tone: 'info',
  },
  {
    label: 'Success rate',
    value: percent(props.value.successRate),
    detail: `${percent(props.value.errorRate)} errors`,
    status: props.value.successRate >= 99 ? 'Healthy' : 'Review errors',
    icon: 'mdi-check-circle-outline',
    tone: props.value.successRate >= 99 ? 'success' : 'warning',
  },
  {
    label: 'Address coverage',
    value: `${number(props.activity.active)} / ${number(props.activity.total)}`,
    detail: `${number(props.activity.passive)} passive`,
    status: props.activity.passive ? 'Action needed' : 'Fully covered',
    icon: 'mdi-access-point-network',
    tone: props.activity.passive ? 'warning' : 'success',
  },
])
</script>

<template>
  <div class="request-summary">
    <v-card v-for="kpi in primary" :key="kpi.label" class="kpi">
      <div class="kpi__heading">
        <v-icon :icon="kpi.icon" :color="kpi.tone" size="20" />
        <small>{{ kpi.label }}</small>
        <span :class="['kpi__status', `kpi__status--${kpi.tone}`]">{{
          kpi.status
        }}</span>
      </div>
      <strong>{{ kpi.value }}</strong>
      <span class="kpi__detail">{{ kpi.detail }}</span>
    </v-card>
  </div>
</template>

<style scoped>
.request-summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}
.kpi {
  position: relative;
  overflow: hidden;
  display: grid;
  gap: 9px;
  min-height: 132px;
  padding: 17px 18px;
  background:
    radial-gradient(
      circle at 100% 0,
      rgba(var(--gfw-flame), 0.1),
      transparent 54%
    ),
    linear-gradient(
      145deg,
      rgb(var(--gfw-surface-strong)),
      rgb(var(--gfw-muted))
    );
  box-shadow: var(--gfw-card-shadow-soft);
}
.kpi::after {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: 3px;
  background: linear-gradient(90deg, #2b211f, #c33c2e, #f7b13b);
  content: '';
  opacity: 0.68;
}
.kpi__heading {
  display: flex;
  align-items: center;
  gap: 8px;
}
.kpi__heading small {
  flex: 1;
  color: rgb(var(--gfw-text-muted));
  font-size: 0.76rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}
.kpi__status {
  padding: 3px 7px;
  border-radius: 999px;
  background: rgba(var(--gfw-ember), 0.08);
  color: rgb(var(--gfw-ember));
  font-size: 0.7rem;
  font-weight: 750;
  white-space: nowrap;
}
.kpi__status--success {
  background: rgba(var(--v-theme-success), 0.12);
  color: rgb(var(--v-theme-success));
}
.kpi__status--warning {
  background: rgba(var(--v-theme-warning), 0.14);
  color: #8a4d00;
}
.kpi > strong {
  color: rgb(var(--gfw-text));
  font-size: clamp(1.65rem, 2.4vw, 2.15rem);
  line-height: 1;
}
.kpi__detail {
  color: rgb(var(--gfw-text-muted));
  font-size: 0.82rem;
}
@media (max-width: 1000px) {
  .request-summary {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (max-width: 560px) {
  .request-summary {
    grid-template-columns: 1fr;
  }
}
</style>
