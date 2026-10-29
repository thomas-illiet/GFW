<script setup lang="ts">
withDefaults(defineProps<{ disabled?: boolean; label?: string }>(), {
  disabled: false,
  label: 'Export CSV',
})
defineEmits<{ export: [] }>()
</script>

<template>
  <v-btn
    variant="tonal"
    color="primary"
    prepend-icon="mdi-download-outline"
    :disabled="disabled"
    @click="$emit('export')"
  >
    {{ label }}
  </v-btn>
</template>
