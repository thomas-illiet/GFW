<script setup lang="ts">
defineProps<{
  title: string
  copy: string
  icon: string
  stat?: string
  statLabel?: string
}>()
</script>
<template>
  <v-card class="header"
    ><div>
      <p>GREAT FIREWALL</p>
      <div class="heading">
        <v-avatar color="primary" variant="tonal" rounded="lg"
          ><v-icon :icon="icon"
        /></v-avatar>
        <div>
          <h1>{{ title }}</h1>
          <span>{{ copy }}</span>
        </div>
      </div>
    </div>
    <div v-if="stat" class="stat">
      <small>{{ statLabel }}</small
      ><strong>{{ stat }}</strong>
    </div></v-card
  >
</template>
<style scoped>
.header {
  position: relative;
  isolation: isolate;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 20px 22px;
  background:
    radial-gradient(
      circle at 92% 0,
      rgba(var(--gfw-flame), 0.14),
      transparent 34%
    ),
    linear-gradient(
      135deg,
      rgb(var(--gfw-surface-strong)),
      rgb(var(--gfw-muted))
    );
  box-shadow: var(--gfw-card-shadow-soft);
}
.header::before {
  position: absolute;
  top: 0;
  left: 24px;
  width: 108px;
  height: 4px;
  border-radius: 0 0 999px 999px;
  background: linear-gradient(90deg, #2b211f, #c33c2e 42%, #f7b13b);
  content: '';
}
.header::after {
  position: absolute;
  z-index: -1;
  top: -150px;
  right: -76px;
  width: 285px;
  height: 285px;
  border: 1px solid rgba(var(--gfw-ember), 0.14);
  border-radius: 50%;
  box-shadow:
    0 0 0 34px rgba(var(--gfw-ember), 0.035),
    0 0 0 68px rgba(var(--gfw-amber), 0.035);
  content: '';
}
.header > div {
  position: relative;
  z-index: 1;
}
p {
  margin: 0 0 10px;
  font: 700 0.7rem var(--gfw-mono);
  letter-spacing: 0.16em;
  color: rgb(var(--gfw-text-muted));
}
.heading {
  display: flex;
  gap: 14px;
  align-items: center;
}
h1 {
  margin: 0;
  color: rgb(var(--gfw-text));
  font-size: clamp(1.5rem, 2.4vw, 2rem);
  font-weight: 800;
  letter-spacing: -0.035em;
}
span {
  color: rgb(var(--gfw-text-secondary));
}
.stat {
  display: grid;
  min-width: 148px;
  padding: 12px 14px;
  border: 1px solid rgba(var(--gfw-ember), 0.2);
  border-radius: 18px;
  background:
    linear-gradient(
      145deg,
      rgba(var(--gfw-ember), 0.09),
      rgba(var(--gfw-amber), 0.04)
    ),
    rgb(var(--gfw-surface-strong));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.7),
    5px 6px 0 rgba(var(--gfw-amber), 0.09);
}
.stat small {
  text-transform: uppercase;
  color: rgb(var(--gfw-text-muted));
}
.stat strong {
  font-size: 1.65rem;
}
@media (max-width: 600px) {
  .header {
    align-items: stretch;
    flex-direction: column;
  }
  .stat {
    min-width: 0;
  }
}
</style>
