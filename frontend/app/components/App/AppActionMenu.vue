<script setup lang="ts">
withDefaults(
  defineProps<{
    ariaLabel?: string
    disabled?: boolean
    disabledHint?: string
  }>(),
  {
    ariaLabel: 'Actions',
    disabled: false,
    disabledHint: '',
  },
)

defineSlots<{
  default(): unknown
}>()
</script>

<template>
  <v-menu location="bottom end">
    <template #activator="{ props: menuProps }">
      <v-tooltip
        :text="disabled && disabledHint ? disabledHint : ''"
        location="top"
      >
        <template #activator="{ props: tooltipProps }">
          <span v-bind="tooltipProps" class="action-menu-activator">
            <v-btn
              v-bind="menuProps"
              class="action-menu-button"
              append-icon="mdi-chevron-down"
              size="small"
              variant="tonal"
              color="primary"
              :disabled="disabled"
              :aria-label="ariaLabel"
            >
              Actions
            </v-btn>
          </span>
        </template>
      </v-tooltip>
    </template>
    <v-list class="action-menu-list" density="compact" min-width="220">
      <slot />
    </v-list>
  </v-menu>
</template>

<style scoped>
.action-menu-button {
  min-width: 104px;
  font-weight: 700;
  letter-spacing: 0;
  text-transform: none;
}
.action-menu-activator {
  display: inline-flex;
}
.action-menu-list {
  padding-block: 6px;
}
</style>
