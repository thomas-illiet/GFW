<script setup lang="ts">
import type { IngestionKey } from '~/types/api'
import { exportCsv } from '~/utils/export'
definePageMeta({ role: 'admin' })
const r = useResources(),
  notifications = useNotifications(),
  dialog = shallowRef(false),
  revokeDialog = shallowRef(false),
  keyToRevoke = shallowRef<IngestionKey | null>(null),
  secret = shallowRef(''),
  search = shallowRef(''),
  showRevoked = shallowRef(false)
const { keys, loading, error } = r
const headers = [
  { title: 'Name', key: 'Name' },
  { title: 'Prefix', key: 'Prefix' },
  { title: 'Last used', key: 'LastUsedAt', align: 'center' as const },
  { title: 'Status', key: 'RevokedAt', align: 'center' as const },
  {
    title: 'Actions',
    key: 'actions',
    align: 'center' as const,
    sortable: false,
  },
]
const filteredKeys = computed(() => {
  const query = search.value.trim().toLowerCase()
  return keys.value.filter(
    (key) =>
      (showRevoked.value || !key.RevokedAt) &&
      (!query || `${key.Name} ${key.Prefix}`.toLowerCase().includes(query)),
  )
})
onMounted(() => void r.loadKeys().catch(() => undefined))
async function create(p: { name: string }) {
  const result = await r.mutation<{ secret: string }>(
    '/api/v1/admin/ingestion-keys',
    'POST',
    p,
  )
  secret.value = result.secret
  dialog.value = false
  await r.loadKeys()
  notifications.success('Ingestion key created')
}
function requestRevoke(key: IngestionKey) {
  keyToRevoke.value = key
  revokeDialog.value = true
}
async function confirmRevoke() {
  if (!keyToRevoke.value) return
  await r.mutation(
    `/api/v1/admin/ingestion-keys/${keyToRevoke.value.ID}`,
    'DELETE',
  )
  revokeDialog.value = false
  keyToRevoke.value = null
  await r.loadKeys()
  notifications.success('Ingestion key revoked')
}
function exportKeys() {
  exportCsv(
    'great-firewall-ingestion-keys.csv',
    ['Name', 'Prefix', 'Status', 'Last used', 'Expires at', 'Created at'],
    filteredKeys.value.map((key) => [
      key.Name,
      key.Prefix,
      key.RevokedAt ? 'revoked' : 'active',
      key.LastUsedAt,
      key.ExpiresAt,
      key.CreatedAt,
    ]),
  )
}
</script>
<template>
  <v-container fluid class="gfw-page"
    ><v-row
      ><v-col cols="12"
        ><AppPageHeader
          title="Ingestion keys"
          copy="Authenticate inventory systems that synchronize managed machine addresses."
          icon="mdi-key-chain"
          :stat="String(keys.filter((k) => !k.RevokedAt).length)"
          stat-label="Active keys" /></v-col
      ><v-col v-if="secret" cols="12"
        ><v-alert
          type="warning"
          title="Copy this secret now"
          closable
          @click:close="secret = ''"
          >It will not be displayed again.
          <div>
            <code class="gfw-mono">{{ secret }}</code>
          </div></v-alert
        ></v-col
      ><v-col cols="12"
        ><KeysIngestionKeyFilters
          v-model:search="search"
          v-model:show-revoked="showRevoked" /></v-col
      ><v-col v-if="error" cols="12"
        ><v-alert type="error" variant="tonal">{{ error }}</v-alert></v-col
      ><v-col cols="12"
        ><AppSectionCard
          title="API keys"
          subtitle="Secrets are hashed and can be revoked at any time."
          icon="mdi-key-chain"
          ><template #actions
            ><AppExportButton
              :disabled="!filteredKeys.length"
              @export="exportKeys"
            /><v-btn
              color="primary"
              prepend-icon="mdi-key-plus"
              @click="dialog = true"
              >Create key</v-btn
            ></template
          ><v-data-table
            class="action-table"
            :headers="headers"
            :items="filteredKeys"
            item-value="ID"
            :items-per-page="-1"
            hide-default-footer
            ><template #[`item.Prefix`]="{ item: k }">
              <code>{{ k.Prefix }}</code>
            </template>
            <template #[`item.LastUsedAt`]="{ item: k }">
              <AppRelativeTime
                :value="k.LastUsedAt"
                prefix="Used"
                empty="Never used"
              />
            </template>
            <template #[`item.RevokedAt`]="{ item: k }">
              <v-chip :color="k.RevokedAt ? 'error' : 'success'" size="small">{{
                k.RevokedAt ? 'Revoked' : 'Active'
              }}</v-chip>
            </template>
            <template #[`item.actions`]="{ item: k }">
              <AppActionMenu
                v-if="!k.RevokedAt"
                :aria-label="`Actions for ${k.Name}`"
              >
                <v-list-item
                  class="text-error"
                  prepend-icon="mdi-key-remove"
                  title="Revoke key"
                  @click="requestRevoke(k)"
                />
              </AppActionMenu>
            </template>
            <template #no-data>
              <AppEmptyState
                icon="mdi-key-off-outline"
                :title="keys.length ? 'No matching keys' : 'No ingestion keys'"
                :text="
                  keys.length
                    ? 'Adjust the search or display revoked keys.'
                    : 'Create a key when an inventory system is ready to synchronize machines.'
                "
              /> </template></v-data-table></AppSectionCard></v-col></v-row
    ><KeysIngestionKeyDialog
      v-model="dialog"
      :loading="loading"
      @save="create" /><KeysIngestionKeyRevokeDialog
      v-model="revokeDialog"
      :ingestion-key="keyToRevoke"
      :loading="loading"
      @confirm="confirmRevoke"
  /></v-container>
</template>
