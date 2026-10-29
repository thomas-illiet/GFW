<script setup lang="ts">
import type { TrustedService } from '~/types/api'

const model = defineModel<boolean>({ required: true })
defineProps<{
  service: TrustedService | null
}>()
</script>

<template>
  <v-dialog v-model="model" max-width="720">
    <v-card class="resolved-dialog" rounded="xl">
      <v-card-title class="resolved-dialog__header">
        <span class="resolved-dialog__icon">
          <v-icon icon="mdi-ip-network-outline" size="28" />
        </span>
        <span>
          <strong>Resolved addresses</strong>
          <small>{{ service?.Name }}</small>
        </span>
        <v-btn
          icon="mdi-close"
          variant="text"
          aria-label="Close resolved addresses"
          @click="model = false"
        />
      </v-card-title>

      <v-card-text class="resolved-dialog__content">
        <div
          v-for="hostname in service?.Hostnames ?? []"
          :key="hostname.ID"
          class="resolved-hostname"
        >
          <div class="resolved-hostname__heading">
            <div>
              <v-icon
                :icon="
                  hostname.Status === 'fresh'
                    ? 'mdi-check-network-outline'
                    : 'mdi-clock-alert-outline'
                "
                :color="hostname.Status === 'fresh' ? 'success' : 'warning'"
              />
              <strong>{{ hostname.Hostname }}</strong>
            </div>
            <v-chip
              :color="hostname.Status === 'fresh' ? 'success' : 'warning'"
              size="small"
              variant="tonal"
            >
              {{ hostname.Status }}
            </v-chip>
          </div>

          <div
            v-if="hostname.IPs.some((ip) => ip.IsCurrent)"
            class="resolved-hostname__addresses"
          >
            <v-chip
              v-for="ip in hostname.IPs.filter((entry) => entry.IsCurrent)"
              :key="ip.ID"
              prepend-icon="mdi-lan-connect"
              variant="outlined"
            >
              {{ ip.IPv4 }}
            </v-chip>
          </div>
          <p v-else class="resolved-hostname__empty">
            No current address resolved.
          </p>
          <p v-if="hostname.LastError" class="resolved-hostname__error">
            <v-icon icon="mdi-alert-circle-outline" size="small" />
            {{ hostname.LastError }}
          </p>
        </div>

        <AppEmptyState
          v-if="!service?.Hostnames.length"
          icon="mdi-dns-outline"
          title="No DNS hostnames"
          text="Add a hostname to this service before resolving addresses."
        />
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.resolved-dialog {
  overflow: hidden;
}
.resolved-dialog__header {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 14px;
  padding: 24px 24px 18px;
}
.resolved-dialog__header > span:nth-child(2) {
  display: grid;
  gap: 2px;
}
.resolved-dialog__header small {
  color: rgb(var(--gfw-text-muted));
  font-size: 0.85rem;
  font-weight: 500;
}
.resolved-dialog__icon {
  display: grid;
  width: 48px;
  height: 48px;
  place-items: center;
  border-radius: 14px;
  color: rgb(var(--v-theme-primary));
  background: rgba(var(--v-theme-primary), 0.1);
}
.resolved-dialog__content {
  display: grid;
  gap: 14px;
  padding: 4px 24px 24px !important;
}
.resolved-hostname {
  display: grid;
  gap: 14px;
  padding: 18px;
  border: 1px solid rgb(var(--gfw-border));
  border-radius: 16px;
  background: rgb(var(--gfw-surface));
}
.resolved-hostname__heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.resolved-hostname__heading > div {
  display: flex;
  align-items: center;
  gap: 9px;
  min-width: 0;
}
.resolved-hostname__heading strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.resolved-hostname__addresses {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.resolved-hostname__empty,
.resolved-hostname__error {
  margin: 0;
  color: rgb(var(--gfw-text-muted));
}
.resolved-hostname__error {
  display: flex;
  align-items: center;
  gap: 6px;
  color: rgb(var(--v-theme-error));
}
@media (max-width: 600px) {
  .resolved-dialog__header {
    padding: 18px 18px 14px;
  }
  .resolved-dialog__content {
    padding: 2px 18px 18px !important;
  }
  .resolved-hostname__heading {
    align-items: flex-start;
  }
}
</style>
