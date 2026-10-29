<script setup lang="ts">
const { mobile } = useVDisplay()
const drawer = shallowRef(!mobile.value)
const drawerRail = shallowRef(false)
const auth = useAuthStore()
const instance = useInstanceStore()
const notifications = useNotifications()
const environment = computed(() => instancePresentation(instance.environment))
const route = useRoute()
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
const currentPage = computed(
  () => links.find((link) => link.to === route.path)?.title ?? 'Dashboard',
)
const notificationIcon = computed(() => {
  const tone = notifications.notification.value?.tone ?? 'success'
  return {
    success: 'mdi-check-circle',
    info: 'mdi-information',
    warning: 'mdi-alert-circle',
    error: 'mdi-alert-octagon',
  }[tone]
})

function toggleNavigation() {
  if (mobile.value) {
    drawer.value = !drawer.value
    return
  }
  drawerRail.value = !drawerRail.value
}

watch(
  mobile,
  (isMobile) => {
    drawer.value = !isMobile
    drawerRail.value = false
  },
  { immediate: true },
)
</script>

<template>
  <v-app>
    <div
      class="gfw-heat-line"
      :style="{ backgroundColor: environment.accent }"
      aria-hidden="true"
    />
    <v-navigation-drawer
      v-model="drawer"
      class="gfw-drawer"
      :rail="drawerRail"
      :rail-width="72"
      :width="288"
      aria-label="Primary navigation"
      floating
    >
      <template #prepend>
        <v-list class="drawer-brand-shell">
          <v-list-item class="drawer-brand" data-test="drawer-brand">
            <template #prepend>
              <span class="drawer-brand__mark" aria-hidden="true">
                <AppFirewallMark :size="30" />
              </span>
            </template>
            <v-list-item-title class="drawer-brand__title">
              Great <span>Firewall</span>
            </v-list-item-title>
            <v-list-item-subtitle class="drawer-brand__subtitle">
              <AppEnvironmentBadge :environment="instance.environment" subtle />
            </v-list-item-subtitle>
          </v-list-item>
        </v-list>
      </template>

      <v-list nav density="compact" class="drawer-navigation">
        <v-list-item
          v-for="link in visibleLinks"
          :key="link.to"
          class="navigation-item"
          :to="link.to"
          :title="link.title"
        >
          <template #prepend>
            <span class="navigation-item__icon" aria-hidden="true">
              <v-icon :icon="link.icon" size="20" />
            </span>
          </template>
        </v-list-item>
      </v-list>

      <template #append>
        <v-list nav density="compact" class="drawer-utility">
          <v-list-item
            class="navigation-item navigation-item--utility"
            to="/documentation"
            title="API Documentation"
          >
            <template #prepend>
              <span class="navigation-item__icon" aria-hidden="true">
                <v-icon icon="mdi-api" size="20" />
              </span>
            </template>
          </v-list-item>
        </v-list>
      </template>
    </v-navigation-drawer>

    <v-app-bar class="gfw-bar" height="64" flat>
      <div class="app-bar__inner">
        <v-app-bar-nav-icon
          aria-label="Toggle navigation"
          rounded="lg"
          class="app-bar__nav"
          data-test="navigation-toggle"
          @click="toggleNavigation"
        />
        <div class="app-bar__context">
          <span class="app-bar__eyebrow">
            <span class="app-bar__beacon" aria-hidden="true" />
            Firewall workspace
          </span>
          <div class="app-bar__breadcrumb">
            <span>Great Firewall</span>
            <v-icon icon="mdi-chevron-right" size="14" />
            <strong>{{ currentPage }}</strong>
          </div>
        </div>
      </div>
      <v-spacer />
      <div class="user">
        <span>{{ auth.user?.name || auth.user?.email }}</span>
        <v-chip size="small" color="primary" variant="tonal">{{
          auth.user?.role
        }}</v-chip>
        <v-btn icon="mdi-logout" aria-label="Log out" @click="auth.logout" />
      </div>
    </v-app-bar>
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
.gfw-heat-line {
  position: fixed;
  z-index: 2500;
  inset: 0 0 auto;
  height: 4px;
  transition: background-color 180ms ease;
}

.gfw-drawer {
  --drawer-rail-inset: 6px;
  overflow: hidden;
  border-right: 1px solid rgba(var(--gfw-border), 0.56);
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
  box-shadow: 0 22px 44px -32px rgba(var(--gfw-shadow), 0.26);
  transition-property: box-shadow, transform, visibility, width, border-radius;
  transition-duration: 220ms;
  transition-timing-function: cubic-bezier(0.2, 0, 0, 1);
}

.gfw-drawer:deep(.v-navigation-drawer__content) {
  overflow-y: auto;
  scrollbar-gutter: stable;
}

.gfw-drawer.v-navigation-drawer--rail {
  border-radius: 0;
}

.gfw-drawer.v-navigation-drawer--rail.v-navigation-drawer--is-hovering {
  border-radius: 0 24px 24px 0;
  box-shadow: 0 20px 48px -30px rgba(var(--gfw-shadow), 0.68);
}

.drawer-brand-shell {
  padding: 10px 9px 6px;
}

.drawer-brand {
  min-height: 66px;
  padding: 8px 10px;
  border: 1px solid rgba(var(--gfw-border), 0.55);
  border-radius: 20px !important;
  background:
    radial-gradient(
      circle at 90% 12%,
      rgba(var(--gfw-flame), 0.1),
      transparent 34%
    ),
    rgba(var(--gfw-surface-strong), 0.76);
  box-shadow: 0 12px 28px -24px rgba(var(--gfw-shadow), 0.58);
  pointer-events: none;
}

.drawer-brand:deep(.v-list-item__overlay) {
  display: none;
}

.drawer-brand__mark {
  display: inline-flex;
  width: 48px;
  height: 48px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  transform: rotate(-4deg);
  border: 1px solid rgba(var(--gfw-ember), 0.16);
  border-radius: 17px;
  background:
    radial-gradient(
      circle at 28% 24%,
      rgba(var(--gfw-amber), 0.46),
      transparent 52%
    ),
    rgba(var(--gfw-ember), 0.08);
  box-shadow: 3px 4px 0 rgba(var(--gfw-flame), 0.1);
  animation: drawer-mark-float 6s ease-in-out 1.2s infinite;
}

.drawer-brand__title {
  color: rgb(var(--gfw-text));
  font-size: 1.2rem;
  font-weight: 800;
  letter-spacing: -0.035em;
  line-height: 1.25;
}

.drawer-brand__title span {
  color: rgb(var(--gfw-ember));
}

.drawer-brand__subtitle {
  display: flex;
  align-items: center;
  margin-top: -2px;
}

.drawer-navigation {
  padding: 4px 9px 18px;
}

.navigation-item {
  position: relative;
  min-height: 48px;
  margin-block: 3px;
  border: 1px solid transparent;
  border-radius: 16px !important;
  color: rgb(var(--gfw-text-muted));
  font-size: 0.87rem;
  font-weight: 700;
  transition:
    background-color 160ms ease,
    border-color 160ms ease,
    transform 160ms ease;
}

.navigation-item__icon {
  display: inline-flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  transform: rotate(-1deg);
  border: 1px solid rgba(var(--gfw-border), 0.52);
  border-radius: 12px;
  background: rgba(var(--gfw-surface-strong), 0.68);
  color: rgb(var(--gfw-text-muted));
  box-shadow: 2px 3px 0 rgba(var(--gfw-amber), 0.08);
  transition:
    transform 160ms ease,
    border-color 160ms ease,
    background-color 160ms ease,
    color 160ms ease;
}

.navigation-item:deep(.v-list-item-title) {
  margin-inline-start: 2px;
}

.navigation-item.v-list-item--active .navigation-item__icon {
  transform: rotate(-4deg);
  border-color: rgba(var(--gfw-ember), 0.24);
  background: rgb(var(--gfw-ember));
  color: rgb(var(--gfw-surface));
  box-shadow: 3px 4px 0 rgba(var(--gfw-flame), 0.14);
}

.navigation-item:not(.v-list-item--active):hover .navigation-item__icon {
  transform: translateY(-1px) rotate(-2deg);
  border-color: rgba(var(--gfw-amber), 0.52);
  color: rgb(var(--gfw-ember));
}

.navigation-item.v-list-item--active {
  position: relative;
  overflow: hidden;
  border: 1px solid rgba(var(--gfw-ember), 0.16);
  background:
    linear-gradient(
      100deg,
      rgba(var(--gfw-ember), 0.15),
      rgba(var(--gfw-amber), 0.11)
    ),
    rgb(var(--gfw-surface-strong));
  box-shadow:
    0 4px 0 -2px rgba(var(--gfw-ember), 0.18),
    0 12px 24px -22px rgba(var(--gfw-shadow), 0.6);
  color: rgb(var(--gfw-text));
}

.navigation-item.v-list-item--active:deep(.v-list-item__overlay) {
  opacity: 0;
}

.navigation-item.v-list-item--active::before {
  position: absolute;
  z-index: 1;
  top: 25%;
  bottom: 25%;
  left: 0;
  width: 3px;
  border-radius: 0 999px 999px 0;
  background: rgb(var(--gfw-ember));
  box-shadow: 0 0 14px rgba(var(--gfw-ember), 0.42);
  content: '';
}

.drawer-utility {
  padding: 8px 9px 10px;
  border-top: 1px solid rgba(var(--gfw-border), 0.48);
  background: rgba(var(--gfw-surface), 0.92);
}

.navigation-item--utility {
  color: rgb(var(--gfw-text-muted));
}

.gfw-drawer.v-navigation-drawer--rail:not(.v-navigation-drawer--is-hovering)
  :deep(.v-list) {
  padding-inline: 0;
}

.gfw-drawer.v-navigation-drawer--rail:not(.v-navigation-drawer--is-hovering)
  :deep(.v-navigation-drawer__content) {
  scrollbar-gutter: auto;
}

.gfw-drawer.v-navigation-drawer--rail:not(.v-navigation-drawer--is-hovering)
  :deep(.v-list-item) {
  position: relative;
  width: calc(100% - (var(--drawer-rail-inset) * 2));
  min-width: 0;
  margin-inline: auto;
  padding-inline: 0;
  grid-template-areas: 'prepend';
  grid-template-columns: minmax(0, 1fr);
  justify-items: center;
  border-radius: 16px !important;
}

.gfw-drawer.v-navigation-drawer--rail:not(.v-navigation-drawer--is-hovering)
  :deep(.v-list-item__prepend) {
  position: absolute;
  inset: 0;
  display: flex;
  width: auto;
  min-width: 0;
  align-items: center;
  justify-content: center;
  margin: 0 !important;
}

.gfw-drawer.v-navigation-drawer--rail:not(.v-navigation-drawer--is-hovering)
  .navigation-item__icon {
  position: static;
  margin: 0;
  transform: rotate(-1deg);
}

.gfw-drawer.v-navigation-drawer--rail:not(.v-navigation-drawer--is-hovering)
  .navigation-item.v-list-item--active
  .navigation-item__icon {
  transform: rotate(-4deg);
}

.gfw-drawer.v-navigation-drawer--rail:not(.v-navigation-drawer--is-hovering)
  .navigation-item.v-list-item--active {
  min-height: 48px;
  border-radius: 16px !important;
}

.gfw-drawer.v-navigation-drawer--rail:not(.v-navigation-drawer--is-hovering)
  .navigation-item.v-list-item--active::after {
  position: absolute;
  z-index: 0;
  inset: 2px 4px 1px 6px;
  transform: translate(-2px, 2px);
  border: 2px solid rgba(var(--gfw-ember), 0.72);
  border-radius: 14px;
  pointer-events: none;
  content: '';
}

.gfw-drawer.v-navigation-drawer--rail:not(.v-navigation-drawer--is-hovering)
  .navigation-item:not(.v-list-item--active):hover
  .navigation-item__icon {
  transform: translateY(-1px) rotate(-2deg);
}

.gfw-drawer.v-navigation-drawer--rail:not(.v-navigation-drawer--is-hovering)
  :deep(.v-list-item__spacer),
.gfw-drawer.v-navigation-drawer--rail:not(.v-navigation-drawer--is-hovering)
  :deep(.v-list-item__content),
.gfw-drawer.v-navigation-drawer--rail:not(.v-navigation-drawer--is-hovering)
  :deep(.v-list-item__append) {
  display: none;
}

.gfw-drawer.v-navigation-drawer--rail:not(.v-navigation-drawer--is-hovering)
  .drawer-brand__mark {
  width: 46px;
  height: 46px;
}

.gfw-bar {
  border-bottom: 1px solid rgba(var(--gfw-border), 0.7) !important;
  background:
    linear-gradient(
      90deg,
      rgba(var(--gfw-amber), 0.1),
      transparent 30%,
      rgba(var(--gfw-flame), 0.08)
    ),
    rgba(var(--gfw-surface), 0.95) !important;
  box-shadow: 0 16px 30px -28px rgba(var(--gfw-shadow), 0.38) !important;
  backdrop-filter: blur(14px);
}

.app-bar__inner {
  display: flex;
  height: 100%;
  align-items: center;
  padding-inline-start: 10px;
}

.app-bar__nav {
  margin-inline-end: 10px;
  border: 1px solid rgba(var(--gfw-amber), 0.44);
  background:
    radial-gradient(
      circle at 28% 24%,
      rgba(var(--gfw-amber), 0.28),
      transparent 55%
    ),
    rgba(var(--gfw-surface-strong), 0.78);
  color: rgb(var(--gfw-text-muted));
  box-shadow: 3px 4px 0 rgba(var(--gfw-flame), 0.08);
}

.app-bar__context {
  display: grid;
  min-width: 0;
  align-content: center;
  gap: 2px;
}

.app-bar__eyebrow {
  display: flex;
  align-items: center;
  gap: 6px;
  padding-inline-start: 4px;
  color: rgb(var(--gfw-text-muted));
  font-family: var(--gfw-mono);
  font-size: 0.58rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1;
  text-transform: uppercase;
}

.app-bar__beacon {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: rgb(var(--gfw-ember));
  box-shadow: 0 0 0 3px rgba(var(--gfw-ember), 0.1);
}

.app-bar__breadcrumb {
  display: flex;
  min-height: 25px;
  align-items: center;
  gap: 4px;
  color: rgb(var(--gfw-text-muted));
  font-size: 0.84rem;
  font-weight: 680;
}

.app-bar__breadcrumb strong {
  color: rgb(var(--gfw-ember));
}

.user {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-right: 8px;
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

@media (max-width: 600px) {
  .app-bar__inner {
    padding-inline-start: 7px;
  }

  .app-bar__nav {
    margin-inline-end: 5px;
  }

  .app-bar__eyebrow,
  .app-bar__breadcrumb span,
  .app-bar__breadcrumb .v-icon {
    display: none;
  }

  .user span {
    display: none;
  }
}

@keyframes drawer-mark-float {
  0%,
  100% {
    transform: rotate(-4deg) translateY(0);
  }
  50% {
    transform: rotate(-2deg) translateY(-2px);
  }
}
</style>
