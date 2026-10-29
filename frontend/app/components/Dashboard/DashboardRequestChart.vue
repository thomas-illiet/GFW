<script setup lang="ts">
import type { Dashboard } from '~/types/api'
import { number } from '~/utils/format'
const props = defineProps<{ items: Readonly<Dashboard['series']> }>()
const chartWidth = 960
const chartHeight = 240
const left = 54
const right = 18
const top = 18
const bottom = 44
const plotWidth = chartWidth - left - right
const plotHeight = chartHeight - top - bottom

const max = computed(() =>
  Math.max(1, ...props.items.map((item) => item.requests)),
)
const x = (index: number) =>
  left + (index / Math.max(1, props.items.length - 1)) * plotWidth
const y = (value: number) => top + plotHeight - (value / max.value) * plotHeight
const requestPoints = computed(() =>
  props.items.map((item, index) => `${x(index)},${y(item.requests)}`).join(' '),
)
const errorPoints = computed(() =>
  props.items.map((item, index) => `${x(index)},${y(item.errors)}`).join(' '),
)
const areaPoints = computed(() =>
  props.items.length
    ? `${left},${top + plotHeight} ${requestPoints.value} ${left + plotWidth},${top + plotHeight}`
    : '',
)
const labelStep = computed(() => Math.max(1, Math.ceil(props.items.length / 6)))
const labels = computed(() =>
  props.items
    .map((item, index) => ({ item, index }))
    .filter(
      ({ index }) =>
        index % labelStep.value === 0 || index === props.items.length - 1,
    ),
)
const totalRequests = computed(() =>
  props.items.reduce((sum, item) => sum + item.requests, 0),
)
const totalErrors = computed(() =>
  props.items.reduce((sum, item) => sum + item.errors, 0),
)
const hoveredIndex = shallowRef<number | null>(null)
const hoveredPoint = computed(() =>
  hoveredIndex.value === null ? null : props.items[hoveredIndex.value],
)
const tooltipStyle = computed(() => {
  if (hoveredIndex.value === null || !hoveredPoint.value) return {}
  return {
    left: `${(x(hoveredIndex.value) / chartWidth) * 100}%`,
    top: `${(y(hoveredPoint.value.requests) / chartHeight) * 100}%`,
  }
})
const tooltipAlignment = computed(() => {
  if (hoveredIndex.value === null) return ''
  const ratio = hoveredIndex.value / Math.max(1, props.items.length - 1)
  if (ratio < 0.18) return 'chart-tooltip--left'
  if (ratio > 0.82) return 'chart-tooltip--right'
  return ''
})
const tooltipPlacement = computed(() => {
  if (!hoveredPoint.value) return ''
  return y(hoveredPoint.value.requests) < 76 ? 'chart-tooltip--below' : ''
})

function label(value: string) {
  const date = new Date(value)
  return props.items.length <= 24
    ? date.toLocaleTimeString('en', { hour: '2-digit', minute: '2-digit' })
    : date.toLocaleDateString('en', { month: 'short', day: 'numeric' })
}
</script>
<template>
  <AppSectionCard
    title="Traffic trend"
    subtitle="Request volume and errors over the selected period"
    icon="mdi-chart-line"
    ><template #actions>
      <div class="chart-legend" aria-label="Chart legend">
        <strong>{{ number(totalRequests) }} requests</strong>
        <span><i class="chart-legend__requests" />Requests</span>
        <span
          ><i class="chart-legend__errors" />{{
            number(totalErrors)
          }}
          errors</span
        >
      </div>
    </template>
    <div v-if="items.length" class="chart">
      <svg
        :viewBox="`0 0 ${chartWidth} ${chartHeight}`"
        role="img"
        aria-label="Firewall requests and errors over time"
      >
        <defs>
          <linearGradient id="request-area" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0" stop-color="#e05a3f" stop-opacity="0.32" />
            <stop offset="1" stop-color="#f7b13b" stop-opacity="0.02" />
          </linearGradient>
        </defs>
        <g class="chart-grid">
          <line
            v-for="ratio in [0, 0.5, 1]"
            :key="ratio"
            :x1="left"
            :x2="left + plotWidth"
            :y1="top + plotHeight * ratio"
            :y2="top + plotHeight * ratio"
          />
        </g>
        <g class="chart-axis">
          <text :x="left - 10" :y="top + 4">{{ number(max) }}</text>
          <text :x="left - 10" :y="top + plotHeight / 2 + 4">
            {{ number(Math.round(max / 2)) }}
          </text>
          <text :x="left - 10" :y="top + plotHeight + 4">0</text>
          <text
            v-for="entry in labels"
            :key="entry.item.at"
            :x="x(entry.index)"
            :y="chartHeight - 12"
            text-anchor="middle"
          >
            {{ label(entry.item.at) }}
          </text>
        </g>
        <polygon :points="areaPoints" fill="url(#request-area)" />
        <polyline
          :points="requestPoints"
          class="chart-line chart-line--requests"
        />
        <polyline :points="errorPoints" class="chart-line chart-line--errors" />
        <g v-for="(item, index) in items" :key="item.at">
          <line
            v-if="hoveredIndex === index"
            :x1="x(index)"
            :x2="x(index)"
            :y1="top"
            :y2="top + plotHeight"
            class="chart-guide"
          />
          <circle
            :cx="x(index)"
            :cy="y(item.requests)"
            r="4"
            class="chart-point chart-point--requests"
          >
            <title>
              {{ label(item.at) }}: {{ item.requests }} requests,
              {{ item.errors }} errors
            </title>
          </circle>
          <circle
            v-if="item.errors"
            :cx="x(index)"
            :cy="y(item.errors)"
            r="3.5"
            class="chart-point chart-point--errors"
          />
          <circle
            :cx="x(index)"
            :cy="y(item.requests)"
            r="12"
            class="chart-hit-area"
            tabindex="0"
            :aria-label="`${label(item.at)}: ${item.requests} requests and ${item.errors} errors`"
            @mouseenter="hoveredIndex = index"
            @mouseleave="hoveredIndex = null"
            @focus="hoveredIndex = index"
            @blur="hoveredIndex = null"
          />
        </g>
      </svg>
      <div
        v-if="hoveredPoint"
        class="chart-tooltip"
        :class="[tooltipAlignment, tooltipPlacement]"
        :style="tooltipStyle"
        role="tooltip"
      >
        <strong>{{ label(hoveredPoint.at) }}</strong>
        <span
          ><i class="chart-tooltip__requests" />{{
            number(hoveredPoint.requests)
          }}
          requests</span
        >
        <span
          ><i class="chart-tooltip__errors" />{{
            number(hoveredPoint.errors)
          }}
          errors</span
        >
      </div>
    </div>
    <AppEmptyState
      v-else
      icon="mdi-chart-line"
      title="No requests yet"
      text="The chart will appear after a firewall retrieves a list."
  /></AppSectionCard>
</template>
<style scoped>
.chart {
  position: relative;
  overflow: visible;
  min-width: 0;
  padding-top: 10px;
}
.chart-guide {
  stroke: rgba(var(--gfw-text), 0.24);
  stroke-dasharray: 3 4;
  stroke-width: 1;
}
.chart-hit-area {
  fill: transparent;
  outline: none;
  cursor: crosshair;
}
.chart-hit-area:focus {
  stroke: rgb(var(--gfw-ember));
  stroke-width: 2;
}
.chart-tooltip {
  --tooltip-x: -50%;
  --tooltip-y: calc(-100% - 10px);

  position: absolute;
  z-index: 2;
  display: grid;
  gap: 5px;
  min-width: 132px;
  padding: 9px 11px;
  border: 1px solid rgba(var(--gfw-ember), 0.25);
  border-radius: 11px;
  background: rgba(43, 33, 31, 0.96);
  box-shadow: 0 8px 20px rgba(var(--gfw-shadow), 0.2);
  color: #fff9f0;
  pointer-events: none;
  transform: translate(var(--tooltip-x), var(--tooltip-y));
}
.chart-tooltip--left {
  --tooltip-x: 0;
}
.chart-tooltip--right {
  --tooltip-x: -100%;
}
.chart-tooltip--below {
  --tooltip-y: 10px;
}
.chart-tooltip strong {
  font-size: 0.76rem;
}
.chart-tooltip span {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.7rem;
  opacity: 0.86;
}
.chart-tooltip i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}
.chart-tooltip__requests {
  background: #f7b13b;
}
.chart-tooltip__errors {
  background: #e05a3f;
}
.chart svg {
  display: block;
  height: auto;
  max-width: 100%;
  width: 100%;
}
.chart-grid line {
  stroke: rgba(var(--gfw-border), 0.55);
  stroke-dasharray: 4 5;
}
.chart-axis text {
  fill: rgb(var(--gfw-text-secondary));
  font-size: 11px;
}
.chart-line {
  fill: none;
  stroke-linecap: round;
  stroke-linejoin: round;
}
.chart-line--requests {
  stroke: rgb(var(--gfw-ember));
  stroke-width: 3;
}
.chart-line--errors {
  stroke: #8e211d;
  stroke-dasharray: 5 5;
  stroke-width: 2;
}
.chart-point {
  stroke: rgb(var(--gfw-surface));
  stroke-width: 2;
}
.chart-point--requests {
  fill: rgb(var(--gfw-flame));
}
.chart-point--errors {
  fill: #8e211d;
}
.chart-legend,
.chart-legend span {
  display: flex;
  align-items: center;
}
.chart-legend {
  gap: 16px;
  color: rgb(var(--gfw-text-secondary));
  font-size: 0.76rem;
}
.chart-legend strong {
  color: rgb(var(--gfw-text));
  font-size: 0.78rem;
}
.chart-legend span {
  gap: 6px;
}
.chart-legend i {
  display: inline-block;
  width: 18px;
  height: 3px;
  border-radius: 999px;
}
.chart-legend__requests {
  background: rgb(var(--gfw-ember));
}
.chart-legend__errors {
  background: #8e211d;
}
</style>
