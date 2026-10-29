<script setup lang="ts">
const open = defineModel<boolean>({ required: true })
const props = withDefaults(
  defineProps<{
    assetType: string
    assetName?: string
    description: string
    impact?: string
    confirmLabel?: string
    icon?: string
    loading?: boolean
  }>(),
  {
    assetName: '',
    impact: 'This action cannot be undone.',
    confirmLabel: 'Delete permanently',
    icon: 'mdi-delete-alert-outline',
    loading: false,
  },
)
const emit = defineEmits<{ confirm: [] }>()

const confirmation = shallowRef('')
const title = computed(() => `Delete ${props.assetType}?`)
const isConfirmed = computed(
  () => Boolean(props.assetName) && confirmation.value === props.assetName,
)

watch([open, () => props.assetName], () => {
  confirmation.value = ''
})

function confirm() {
  if (isConfirmed.value && !props.loading) emit('confirm')
}
</script>

<template>
  <v-dialog v-model="open" max-width="580" persistent>
    <v-card class="delete-dialog" rounded="xl">
      <v-card-text class="pa-7">
        <header class="delete-dialog__header">
          <div class="delete-dialog__icon" aria-hidden="true">
            <v-icon :icon="icon" size="30" />
          </div>
          <div>
            <p class="delete-dialog__eyebrow">Permanent deletion</p>
            <h2>{{ title }}</h2>
            <p class="delete-dialog__description">{{ description }}</p>
          </div>
        </header>

        <div class="delete-dialog__impact">
          <v-icon icon="mdi-alert-circle-outline" size="20" />
          <div>
            <strong>This action cannot be undone</strong>
            <span>{{ impact }}</span>
          </div>
        </div>

        <div class="delete-dialog__confirmation">
          <label for="delete-asset-confirmation">
            Type <code>{{ assetName }}</code> to confirm
          </label>
          <v-text-field
            id="delete-asset-confirmation"
            v-model="confirmation"
            :placeholder="assetName"
            prepend-inner-icon="mdi-form-textbox"
            variant="outlined"
            density="comfortable"
            autocomplete="off"
            autofocus
            hide-details
            @keydown.enter.prevent="confirm"
          />
          <small v-if="confirmation && !isConfirmed">
            The name must match exactly.
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
          prepend-icon="mdi-delete-outline"
          :loading="loading"
          :disabled="!isConfirmed"
          @click="confirm"
        >
          {{ confirmLabel }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.delete-dialog {
  border: 1px solid rgba(var(--v-theme-error), 0.2);
}
.delete-dialog__header {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 18px;
}
.delete-dialog__icon {
  display: grid;
  width: 56px;
  height: 56px;
  place-items: center;
  border-radius: 17px;
  background: rgba(var(--v-theme-error), 0.12);
  color: rgb(var(--v-theme-error));
}
.delete-dialog__eyebrow {
  margin: 0 0 4px;
  color: rgb(var(--v-theme-error));
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.delete-dialog h2 {
  margin: 0;
  color: rgb(var(--gfw-text));
  font-size: 1.45rem;
}
.delete-dialog__description {
  margin: 8px 0 0;
  color: rgb(var(--gfw-text-secondary));
  line-height: 1.55;
}
.delete-dialog__impact {
  display: flex;
  gap: 12px;
  margin: 22px 0;
  padding: 14px 16px;
  border-radius: 14px;
  background: rgba(var(--v-theme-error), 0.08);
  color: rgb(var(--v-theme-error));
}
.delete-dialog__impact div {
  display: grid;
  gap: 2px;
}
.delete-dialog__impact span {
  color: rgb(var(--gfw-text-secondary));
  font-size: 0.82rem;
}
.delete-dialog__confirmation {
  display: grid;
  gap: 9px;
}
.delete-dialog__confirmation label {
  color: rgb(var(--gfw-text-secondary));
  font-size: 0.9rem;
}
.delete-dialog__confirmation code {
  color: rgb(var(--gfw-text));
  font-weight: 700;
}
.delete-dialog__confirmation small {
  color: rgb(var(--v-theme-error));
}

@media (max-width: 600px) {
  .delete-dialog__header {
    grid-template-columns: 1fr;
  }
}
</style>
