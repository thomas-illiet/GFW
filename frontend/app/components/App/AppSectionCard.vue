<script setup lang="ts">
defineProps<{ title: string; subtitle?: string; icon?: string }>()
defineSlots<{ actions(): unknown; default(): unknown }>()
</script>
<template>
  <v-card class="section"
    ><header>
      <div>
        <h2><v-icon v-if="icon" :icon="icon" color="primary" /> {{ title }}</h2>
        <p v-if="subtitle">{{ subtitle }}</p>
      </div>
      <div class="actions"><slot name="actions" /></div>
    </header>
    <slot
  /></v-card>
</template>
<style scoped>
.section {
  min-width: 0;
  padding: 20px;
  background:
    radial-gradient(
      circle at 92% 4%,
      rgba(var(--gfw-flame), 0.05),
      transparent 32%
    ),
    linear-gradient(
      180deg,
      rgb(var(--gfw-surface-strong)),
      rgb(var(--gfw-surface)) 54%,
      rgb(var(--gfw-muted)) 130%
    );
  box-shadow: var(--gfw-card-shadow-soft);
}
header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}
header > div:first-child {
  min-width: 0;
}
h2 {
  margin: 0;
  font-size: 1.12rem;
}
p {
  margin: 4px 0 0;
  color: rgb(var(--gfw-text-secondary));
}
.actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  flex-wrap: wrap;
}
@media (max-width: 1100px) {
  header {
    align-items: stretch;
    flex-direction: column;
  }
  .actions {
    justify-content: flex-start;
  }
}
@media (max-width: 600px) {
  .section {
    padding: 16px;
  }
  .actions :deep(.v-btn) {
    width: 100%;
    flex: 1 1 100%;
  }
}
</style>
