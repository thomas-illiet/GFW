<script setup lang="ts">
import type { DeepReadonly } from 'vue'

import type { Dashboard } from '~/types/api'
import { number, percent } from '~/utils/format'

const props = defineProps<{
  formats: DeepReadonly<Dashboard['formats']>
  sources: DeepReadonly<Dashboard['sources']>
}>()

const groups = computed(() => [
  {
    label: 'Response formats',
    icon: 'mdi-code-json',
    items: props.formats,
    total: props.formats.reduce((sum, item) => sum + item.count, 0),
    color: 'info',
  },
  {
    label: 'Address sources',
    icon: 'mdi-source-branch',
    items: props.sources,
    total: props.sources.reduce((sum, item) => sum + item.count, 0),
    color: 'primary',
  },
])

function share(count: number, total: number) {
  return total ? (count * 100) / total : 0
}
</script>

<template>
  <AppSectionCard
    title="Traffic mix"
    subtitle="Relative usage by format and source"
    icon="mdi-chart-donut"
  >
    <div v-if="formats.length || sources.length" class="mix-groups">
      <section v-for="group in groups" :key="group.label" class="mix-group">
        <h3><v-icon :icon="group.icon" size="18" />{{ group.label }}</h3>
        <div v-for="item in group.items" :key="item.name" class="mix-row">
          <div class="mix-row__label">
            <strong>{{ item.name.toUpperCase() }}</strong>
            <span>{{ number(item.count) }}</span>
          </div>
          <v-progress-linear
            :model-value="share(item.count, group.total)"
            :color="group.color"
            height="10"
            rounded
          />
          <small>{{ percent(share(item.count, group.total)) }}</small>
        </div>
      </section>
    </div>
    <AppEmptyState
      v-else
      icon="mdi-chart-donut"
      title="No traffic mix yet"
      text="Formats and sources will be summarized after the first request."
    />
  </AppSectionCard>
</template>

<style scoped>
.mix-groups,
.mix-group,
.mix-row {
  display: grid;
}
.mix-groups {
  gap: 22px;
}
.mix-group {
  gap: 12px;
}
.mix-group + .mix-group {
  padding-top: 20px;
  border-top: 1px solid rgba(var(--gfw-border), 0.5);
}
.mix-group h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  color: rgb(var(--gfw-text-secondary));
  font-size: 0.78rem;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}
.mix-row {
  grid-template-columns: 1fr 52px;
  align-items: center;
  gap: 7px 10px;
}
.mix-row__label {
  grid-column: 1 / 3;
  display: flex;
  justify-content: space-between;
  font-size: 0.78rem;
}
.mix-row__label span,
.mix-row small {
  color: rgb(var(--gfw-text-secondary));
}
.mix-row small {
  text-align: right;
}
</style>
