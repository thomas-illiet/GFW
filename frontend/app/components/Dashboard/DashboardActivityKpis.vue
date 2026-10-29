<script setup lang="ts">
import type { Dashboard } from '~/types/api'
import { number, percent } from '~/utils/format'

const props = defineProps<{ value: Dashboard['activity'] }>()

const groups = computed(() => [
  {
    label: 'Machines',
    active: props.value.machineActive,
    total: props.value.machineActive + props.value.machinePassive,
    icon: 'mdi-server-network',
    color: 'info',
  },
  {
    label: 'Trusted',
    active: props.value.trustedActive,
    total: props.value.trustedActive + props.value.trustedPassive,
    icon: 'mdi-shield-check-outline',
    color: 'success',
  },
])

const coverageStyle = computed(() => ({
  background: `conic-gradient(rgb(var(--gfw-ember)) ${Math.min(100, props.value.activeRate)}%, rgba(var(--gfw-border), .35) 0)`,
}))

function rate(active: number, total: number) {
  return total ? (active * 100) / total : 0
}
</script>

<template>
  <AppSectionCard
    title="Address coverage"
    subtitle="Addresses served in this period"
    icon="mdi-radar"
  >
    <div class="coverage-overview">
      <div class="coverage-ring" :style="coverageStyle">
        <div>
          <strong>{{ percent(value.activeRate) }}</strong
          ><span>active</span>
        </div>
      </div>
      <div class="coverage-copy">
        <strong>{{ number(value.active) }} of {{ number(value.total) }}</strong>
        <span>distinct addresses were served</span>
        <v-chip
          :color="value.passive ? 'warning' : 'success'"
          size="small"
          variant="tonal"
          :prepend-icon="
            value.passive ? 'mdi-alert-outline' : 'mdi-check-circle-outline'
          "
        >
          {{
            value.passive
              ? `${number(value.passive)} need attention`
              : 'Full coverage'
          }}
        </v-chip>
      </div>
    </div>
    <div class="coverage-groups">
      <div v-for="group in groups" :key="group.label" class="coverage-group">
        <div>
          <v-icon :icon="group.icon" :color="group.color" size="18" />
          <span>{{ group.label }}</span>
          <strong>{{ group.active }}/{{ group.total }}</strong>
        </div>
        <v-progress-linear
          :model-value="rate(group.active, group.total)"
          :color="group.color"
          height="7"
          rounded
        />
      </div>
    </div>
  </AppSectionCard>
</template>

<style scoped>
.coverage-overview {
  display: grid;
  grid-template-columns: 116px 1fr;
  align-items: center;
  gap: 18px;
  margin-bottom: 22px;
}
.coverage-ring {
  display: grid;
  width: 116px;
  height: 116px;
  padding: 11px;
  border-radius: 50%;
  place-items: center;
}
.coverage-ring > div {
  display: grid;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: rgb(var(--gfw-surface-strong));
  place-content: center;
  text-align: center;
}
.coverage-ring strong {
  font-size: 1.35rem;
  line-height: 1;
}
.coverage-ring span,
.coverage-copy span {
  color: rgb(var(--gfw-text-secondary));
}
.coverage-ring span {
  margin-top: 4px;
  font-size: 0.68rem;
  text-transform: uppercase;
}
.coverage-copy {
  display: grid;
  justify-items: start;
  gap: 7px;
}
.coverage-copy > strong {
  font-size: 1.2rem;
}
.coverage-copy > span {
  font-size: 0.78rem;
  line-height: 1.4;
}
.coverage-groups {
  display: grid;
  gap: 15px;
}
.coverage-group {
  display: grid;
  gap: 7px;
}
.coverage-group > div {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 8px;
}
.coverage-group span {
  color: rgb(var(--gfw-text-secondary));
  font-size: 0.8rem;
}
.coverage-group strong {
  font-size: 0.82rem;
}
@media (max-width: 440px) {
  .coverage-overview {
    grid-template-columns: 1fr;
    justify-items: center;
    text-align: center;
  }
  .coverage-copy {
    justify-items: center;
  }
}
</style>
