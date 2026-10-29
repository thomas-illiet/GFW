<script setup lang="ts">
defineProps<{
  icon: string
  title: string
  text: string
  actionLabel?: string
  actionIcon?: string
}>()
defineEmits<{ action: [] }>()
</script>
<template>
  <div class="empty">
    <v-icon :icon="icon" size="48" color="primary" /><strong>{{ title }}</strong
    ><span>{{ text }}</span
    ><v-btn
      v-if="actionLabel"
      class="mt-2"
      color="primary"
      variant="tonal"
      :prepend-icon="actionIcon"
      @click="$emit('action')"
      >{{ actionLabel }}</v-btn
    >
  </div>
</template>
<style scoped>
.empty {
  display: grid;
  justify-items: center;
  gap: 8px;
  padding: 36px 20px;
  color: rgb(var(--gfw-text-secondary));
}
.empty strong {
  color: rgb(var(--gfw-text));
  font-size: 1rem;
}
</style>
