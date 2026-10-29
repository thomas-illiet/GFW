<script setup lang="ts">
const { mdAndUp } = useVDisplay()
const drawer = shallowRef(mdAndUp.value)
const auth = useAuthStore()
const notifications = useNotifications()
const links = [
  { title: 'Dashboard', icon: 'mdi-view-dashboard-outline', to: '/' },
  { title: 'Machines', icon: 'mdi-server-network', to: '/machines' },
  { title: 'Ecosystems', icon: 'mdi-cube-outline', to: '/ecosystems' },
  {
    title: 'Trusted services',
    icon: 'mdi-shield-check-outline',
    to: '/trusted-services',
  },
  {
    title: 'Effective lists',
    icon: 'mdi-format-list-checks',
    to: '/effective-lists',
  },
  {
    title: 'Ingestion keys',
    icon: 'mdi-key-chain',
    to: '/ingestion-keys',
    admin: true,
  },
]
const visibleLinks = computed(() =>
  links.filter((link) => !link.admin || auth.can('admin')),
)
watch(mdAndUp, (isDesktop) => {
  drawer.value = isDesktop
})
const notificationIcon = computed(() => {
  const tone = notifications.notification.value?.tone ?? 'success'
  return {
    success: 'mdi-check-circle',
    info: 'mdi-information',
    warning: 'mdi-alert-circle',
    error: 'mdi-alert-octagon',
  }[tone]
})
</script>
<template>
  <v-app>
    <div class="gfw-heat-line" aria-hidden="true" />
    <v-navigation-drawer v-model="drawer" class="gfw-drawer" width="276">
      <div class="brand">
        <AppFirewallMark :size="46" />
        <div>
          <strong>Great Firewall</strong><small>Policy control plane</small>
        </div>
      </div>
      <v-list nav class="px-3"
        ><v-list-item
          v-for="link in visibleLinks"
          :key="link.to"
          :to="link.to"
          :title="link.title"
          :prepend-icon="link.icon"
          rounded="lg"
      /></v-list>
    </v-navigation-drawer>
    <v-app-bar class="gfw-bar" flat
      ><v-app-bar-nav-icon @click="drawer = !drawer" /><v-app-bar-title
        >Great Firewall</v-app-bar-title
      ><v-spacer />
      <div class="user">
        <span>{{ auth.user?.name || auth.user?.email }}</span
        ><v-chip size="small" color="primary" variant="tonal">{{
          auth.user?.role
        }}</v-chip
        ><v-btn
          icon="mdi-logout"
          aria-label="Log out"
          @click="auth.logout"
        /></div
    ></v-app-bar>
    <v-main><slot /></v-main>
    <v-snackbar
      class="app-snackbar"
      :model-value="Boolean(notifications.notification.value)"
      :color="notifications.notification.value?.tone"
      location="bottom right"
      rounded="lg"
      timeout="4000"
      @update:model-value="(value) => !value && notifications.clear()"
    >
      <div class="notification-content">
        <v-icon :icon="notificationIcon" size="22" />
        <span>{{ notifications.notification.value?.message }}</span>
      </div>
      <template #actions>
        <v-btn
          class="notification-close"
          color="white"
          icon="mdi-close"
          size="small"
          variant="text"
          aria-label="Dismiss"
          @click="notifications.clear"
        />
      </template>
    </v-snackbar>
  </v-app>
</template>
<style scoped>
.gfw-drawer {
  background:
    radial-gradient(
      circle at 16% 92%,
      rgba(var(--gfw-amber), 0.16),
      transparent 28%
    ),
    radial-gradient(
      circle at 88% 8%,
      rgba(var(--gfw-flame), 0.11),
      transparent 24%
    ),
    linear-gradient(
      180deg,
      rgba(var(--gfw-surface), 0.99),
      rgba(var(--gfw-muted), 0.99) 76%,
      rgba(var(--gfw-ember), 0.09) 140%
    );
  border-right: 1px solid rgba(var(--gfw-border), 0.56);
  box-shadow: 0 22px 44px -32px rgba(var(--gfw-shadow), 0.26);
}
.gfw-bar {
  background:
    linear-gradient(
      90deg,
      rgba(var(--gfw-amber), 0.1),
      transparent 30%,
      rgba(var(--gfw-flame), 0.08)
    ),
    rgba(var(--gfw-surface), 0.95) !important;
  border-bottom: 1px solid rgba(var(--gfw-border), 0.7) !important;
  box-shadow: 0 16px 30px -28px rgba(var(--gfw-shadow), 0.38) !important;
  backdrop-filter: blur(14px);
}
.gfw-heat-line {
  position: fixed;
  z-index: 2500;
  inset: 0 0 auto;
  height: 4px;
  background: linear-gradient(
    90deg,
    #2b211f 0 12%,
    #c33c2e 35%,
    #f7b13b 72%,
    #e05a3f
  );
}
.app-snackbar :deep(.v-snackbar__wrapper) {
  min-width: 0 !important;
  width: auto !important;
  max-width: min(420px, calc(100vw - 32px)) !important;
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 14px 36px -18px rgba(var(--gfw-shadow), 0.64) !important;
}
.app-snackbar :deep(.v-snackbar__content) {
  padding: 11px 8px 11px 14px !important;
}
.app-snackbar :deep(.v-snackbar__actions) {
  margin-inline-end: 5px;
}
.notification-content {
  display: flex;
  min-width: 230px;
  align-items: center;
  gap: 10px;
}
.notification-content span {
  font-size: 0.88rem;
  font-weight: 650;
  line-height: 1.35;
}
.notification-close {
  opacity: 0.82;
}
.notification-close:hover {
  opacity: 1;
}
.gfw-drawer :deep(.v-list-item--active) {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(var(--gfw-ember), 0.18);
  background:
    linear-gradient(
      100deg,
      rgba(var(--gfw-ember), 0.15),
      rgba(var(--gfw-amber), 0.12)
    ),
    rgb(var(--gfw-surface-strong));
  box-shadow:
    0 4px 0 -2px rgba(var(--gfw-ember), 0.2),
    0 12px 24px -22px rgba(var(--gfw-shadow), 0.65);
  color: rgb(var(--gfw-text)) !important;
}
.gfw-drawer :deep(.v-list-item--active .v-list-item__overlay) {
  opacity: 0;
}
.gfw-drawer :deep(.v-list-item--active .v-icon) {
  color: rgb(var(--gfw-ember));
}
.gfw-drawer :deep(.v-list-item--active)::before {
  position: absolute;
  inset: 25% auto 25% 0;
  width: 3px;
  border-radius: 0 999px 999px 0;
  background: rgb(var(--gfw-ember));
  box-shadow: 0 0 14px rgba(var(--gfw-ember), 0.48);
  content: '';
}
.gfw-drawer :deep(.v-list-item:hover) {
  background: rgba(var(--gfw-amber), 0.12);
}
.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 22px 18px 18px;
}
.brand div {
  display: grid;
}
.brand strong {
  color: rgb(var(--gfw-text));
  font-size: 1.08rem;
  letter-spacing: -0.02em;
}
.brand small {
  color: rgb(var(--gfw-text-muted));
}
.user {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-right: 8px;
}
@media (max-width: 600px) {
  .user span {
    display: none;
  }
}
</style>
