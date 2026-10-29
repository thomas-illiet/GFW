<script setup lang="ts">
definePageMeta({ layout: 'public', auth: false })

const auth = useAuthStore()
const pending = shallowRef(false)

function signIn() {
  pending.value = true
  auth.login()
}
</script>

<template>
  <v-container fluid class="login-page">
    <div class="login-composition">
      <div class="login-composition__heat-line" aria-hidden="true" />
      <LoginVisual status-label="Perimeter online" status-tone="online" />
      <LoginPanel :pending="pending" @sign-in="signIn" />
    </div>
  </v-container>
</template>

<style scoped>
.login-page {
  display: grid;
  min-height: 100vh;
  place-items: center;
  padding: clamp(16px, 4vw, 52px);
}
.login-composition {
  position: relative;
  display: grid;
  width: min(1180px, 100%);
  min-height: min(660px, calc(100vh - clamp(32px, 8vw, 104px)));
  grid-template-columns: minmax(0, 1.16fr) minmax(390px, 0.84fr);
  overflow: hidden;
  border: 2px solid rgba(var(--gfw-soot), 0.2);
  border-radius: clamp(24px, 3vw, 38px);
  background: rgb(var(--gfw-surface));
  box-shadow:
    12px 14px 0 rgba(var(--gfw-ember), 0.13),
    -8px -8px 0 rgba(var(--gfw-amber), 0.13),
    0 44px 92px -48px rgba(var(--gfw-shadow), 0.62);
}
.login-composition__heat-line {
  position: absolute;
  z-index: 5;
  top: 0;
  right: 0;
  left: 0;
  height: 6px;
  background: linear-gradient(90deg, #2b211f, #c33c2e 42%, #f7b13b);
}
@media (max-width: 960px) {
  .login-composition {
    min-height: 0;
    grid-template-columns: 1fr;
  }
}
@media (max-width: 600px) {
  .login-page {
    align-items: start;
    padding: 10px;
  }
  .login-composition {
    border-width: 1px;
    border-radius: 24px;
    box-shadow:
      5px 7px 0 rgba(var(--gfw-ember), 0.11),
      0 24px 56px -38px rgba(var(--gfw-shadow), 0.42);
  }
}
</style>
