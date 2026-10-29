<script setup lang="ts">
defineProps<{ error?: string | null; pending: boolean }>()
defineEmits<{ signIn: [] }>()
</script>

<template>
  <v-card class="login-panel" rounded="0">
    <v-card-text class="login-panel__content pa-7 pa-sm-9 pa-md-10">
      <div class="login-brand">
        <AppFirewallMark :size="58" />
        <div class="login-brand__copy">
          <span>Great <b>Firewall</b></span>
          <small>Network policy control plane</small>
        </div>
      </div>
      <div class="login-copy">
        <p class="login-kicker">Welcome to the perimeter</p>
        <h1>Keep the right routes open.</h1>
        <p class="login-description">
          Manage machine addresses, trusted services and the effective lists
          consumed by your firewalls from one secure workspace.
        </p>
      </div>
      <div class="login-panel__breathing-room" aria-hidden="true" />
      <v-alert v-if="error" type="warning" variant="tonal" rounded="lg">
        {{ error }}
      </v-alert>
      <v-btn
        block
        height="48"
        size="large"
        color="primary"
        rounded="xl"
        class="login-action text-none"
        prepend-icon="mdi-login-variant"
        append-icon="mdi-arrow-right"
        :loading="pending"
        @click="$emit('signIn')"
      >
        Enter Great Firewall
      </v-btn>
      <div class="login-assurance">
        <v-icon icon="mdi-shield-lock-outline" size="17" />
        <span>Authentication stays with your organization.</span>
      </div>
    </v-card-text>
  </v-card>
</template>

<style scoped>
.login-panel {
  position: relative;
  height: 100%;
  border: 0;
  background:
    radial-gradient(circle at 92% 8%, rgba(247, 177, 59, 0.2), transparent 24%),
    linear-gradient(
      160deg,
      rgb(var(--gfw-surface-strong)),
      rgb(var(--gfw-surface)) 62%,
      rgba(var(--gfw-ember), 0.08)
    );
  box-shadow: none;
}
.login-panel::after {
  position: absolute;
  top: 25px;
  right: 28px;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: rgb(var(--gfw-ember));
  box-shadow:
    20px 12px 0 -2px #f7b13b,
    5px 28px 0 -4px rgba(var(--gfw-soot), 0.48);
  content: '';
  opacity: 0.75;
}
.login-panel__content {
  display: flex;
  height: 100%;
  flex-direction: column;
  gap: 22px;
}
.login-brand {
  display: flex;
  align-items: center;
  gap: 14px;
}
.login-brand__copy {
  display: grid;
  gap: 2px;
}
.login-brand__copy > span {
  color: rgb(var(--gfw-text));
  font-size: 1.35rem;
  font-weight: 850;
  letter-spacing: -0.03em;
}
.login-brand__copy b {
  color: rgb(var(--gfw-ember));
}
.login-brand__copy small {
  color: rgb(var(--gfw-text-muted));
  font-size: 0.7rem;
  font-weight: 750;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.login-copy {
  display: grid;
  gap: 12px;
  margin-top: auto;
}
.login-kicker {
  margin: 0;
  color: rgb(var(--gfw-ember));
  font: 850 0.74rem var(--gfw-mono);
  letter-spacing: 0.13em;
  text-transform: uppercase;
}
.login-copy h1 {
  margin: 0;
  color: rgb(var(--gfw-text));
  font-size: clamp(2.2rem, 3.4vw, 3.35rem);
  font-weight: 900;
  line-height: 0.98;
  letter-spacing: -0.055em;
}
.login-description {
  margin: 0;
  max-width: 29rem;
  color: rgb(var(--gfw-text-secondary));
  font-size: 1.02rem;
  line-height: 1.65;
}
.login-panel__breathing-room {
  height: 96px;
  flex: 0 0 96px;
}
.login-action {
  height: 48px;
  min-height: 48px;
  max-height: 48px;
  flex: 0 0 48px;
  margin-top: auto;
}
.login-assurance {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 7px;
  color: rgb(var(--gfw-text-muted));
  font-size: 0.78rem;
  text-align: center;
}
@media (max-width: 960px) {
  .login-copy,
  .login-action {
    margin-top: 0;
  }
  .login-panel__breathing-room {
    display: none;
  }
}
@media (max-width: 600px) {
  .login-panel::after {
    transform: scale(0.8);
  }
  .login-brand__copy > span {
    font-size: 1.2rem;
  }
  .login-copy h1 {
    font-size: 2.2rem;
  }
}
</style>
