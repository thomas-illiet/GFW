<script setup lang="ts">
defineProps<{ icon: string; title: string; text: string }>()
</script>
<template>
  <div class="empty">
    <v-icon :icon="icon" size="48" color="primary" /><strong>{{ title }}</strong
    ><span>{{ text }}</span>
  </div>
</template>
<style scoped>
.empty {
  display: grid;
  justify-items: center;
  gap: 8px;
  padding: 48px 20px;
  color: rgb(var(--gfw-text-secondary));
}
.empty strong {
  color: rgb(var(--gfw-text));
  font-size: 1.05rem;
}
</style>
