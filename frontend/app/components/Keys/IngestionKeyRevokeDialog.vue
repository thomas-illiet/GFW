<script setup lang="ts">
import type { IngestionKey } from '~/types/api'

const open = defineModel<boolean>({ required: true })
const props = defineProps<{
  ingestionKey: IngestionKey | null
  loading: boolean
}>()
const emit = defineEmits<{ confirm: [] }>()

const confirmation = shallowRef('')
const expectedName = computed(() => props.ingestionKey?.Name ?? '')
const isConfirmed = computed(
  () =>
    Boolean(expectedName.value) && confirmation.value === expectedName.value,
)

watch([open, expectedName], () => {
  confirmation.value = ''
})

function confirm() {
  if (isConfirmed.value && !props.loading) emit('confirm')
}
</script>

<template>
  <v-dialog v-model="open" max-width="560" persistent>
    <v-card class="revoke-dialog" rounded="xl">
      <v-card-text class="pa-7">
        <header class="revoke-dialog__header">
          <div class="revoke-dialog__icon" aria-hidden="true">
            <v-icon icon="mdi-key-alert-outline" size="30" />
          </div>
          <div>
            <p class="revoke-dialog__eyebrow">Permanent access change</p>
            <h2>Revoke ingestion key?</h2>
            <p class="revoke-dialog__description">
              The inventory system using this key will immediately lose access
              to machine ingestion.
            </p>
          </div>
        </header>

        <div class="revoke-dialog__impact">
          <v-icon icon="mdi-alert-circle-outline" size="20" />
          <div>
            <strong>This action cannot be undone</strong>
            <span>A new key must be created to restore access.</span>
          </div>
        </div>

        <div class="revoke-dialog__confirmation">
          <label for="revoke-key-name">
            Type <code>{{ expectedName }}</code> to confirm
          </label>
          <v-text-field
            id="revoke-key-name"
            v-model="confirmation"
            :placeholder="expectedName"
            prepend-inner-icon="mdi-form-textbox"
            variant="outlined"
            density="comfortable"
            autocomplete="off"
            autofocus
            hide-details
            @keydown.enter.prevent="confirm"
          />
          <small v-if="confirmation && !isConfirmed">
            The key name must match exactly.
          </small>
        </div>
      </v-card-text>

      <v-card-actions class="px-7 pb-7 pt-0">
        <v-spacer />
        <v-btn variant="text" :disabled="loading" @click="open = false">
          Cancel
        </v-btn>
        <v-btn
          color="error"
          variant="flat"
          prepend-icon="mdi-key-remove"
          :loading="loading"
          :disabled="!isConfirmed"
          @click="confirm"
        >
          Revoke key
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.revoke-dialog {
  border: 1px solid rgba(var(--v-theme-error), 0.2);
}
.revoke-dialog__header {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 18px;
}
.revoke-dialog__icon {
  display: grid;
  width: 56px;
  height: 56px;
  place-items: center;
  border-radius: 17px;
  background: rgba(var(--v-theme-error), 0.12);
  color: rgb(var(--v-theme-error));
}
.revoke-dialog__eyebrow {
  margin: 0 0 4px;
  color: rgb(var(--v-theme-error));
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.revoke-dialog h2 {
  margin: 0;
  color: rgb(var(--gfw-text));
  font-size: 1.45rem;
}
.revoke-dialog__description {
  margin: 8px 0 0;
  color: rgb(var(--gfw-text-secondary));
  line-height: 1.55;
}
.revoke-dialog__impact {
  display: flex;
  gap: 12px;
  margin: 22px 0;
  padding: 14px 16px;
  border-radius: 14px;
  background: rgba(var(--v-theme-error), 0.08);
  color: rgb(var(--v-theme-error));
}
.revoke-dialog__impact div {
  display: grid;
  gap: 2px;
}
.revoke-dialog__impact span {
  color: rgb(var(--gfw-text-secondary));
  font-size: 0.82rem;
}
.revoke-dialog__confirmation {
  display: grid;
  gap: 9px;
}
.revoke-dialog__confirmation label {
  color: rgb(var(--gfw-text-secondary));
  font-size: 0.9rem;
}
.revoke-dialog__confirmation code {
  color: rgb(var(--gfw-text));
  font-weight: 700;
}
.revoke-dialog__confirmation small {
  color: rgb(var(--v-theme-error));
}

@media (max-width: 600px) {
  .revoke-dialog__header {
    grid-template-columns: 1fr;
  }
}
</style>
