<script setup lang="ts">
import { dateTime, relativeTime } from '~/utils/format'

const props = withDefaults(
  defineProps<{
    value?: string | null
    prefix?: string
    empty?: string
  }>(),
  {
    value: null,
    prefix: '',
    empty: 'Never',
  },
)

const label = computed(() => {
  if (!props.value) return props.empty
  const relative = relativeTime(props.value)
  return props.prefix ? `${props.prefix} ${relative}` : relative
})

const exact = computed(() =>
  props.value ? dateTime(props.value) : props.empty,
)
</script>

<template>
  <span class="relative-time" :title="exact">
    <v-icon icon="mdi-clock-outline" size="15" aria-hidden="true" />
    <span>{{ label }}</span>
  </span>
</template>

<style scoped>
.relative-time {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: rgb(var(--gfw-text-secondary));
  font-size: 0.82rem;
  white-space: nowrap;
}
</style>
