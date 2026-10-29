<script setup lang="ts">
withDefaults(
  defineProps<{
    title: string
    text: string
    subject?: string
    confirmLabel?: string
    confirmIcon?: string
    confirmColor?: string
    icon?: string
    loading?: boolean
  }>(),
  {
    subject: '',
    confirmLabel: 'Delete',
    confirmIcon: 'mdi-delete-outline',
    confirmColor: 'error',
    icon: 'mdi-alert-outline',
    loading: false,
  },
)

const model = defineModel<boolean>({ required: true })
defineEmits<{ confirm: [] }>()
</script>

<template>
  <v-dialog v-model="model" max-width="480" persistent>
    <v-card rounded="xl">
      <v-card-text class="confirm-dialog pa-7">
        <div
          class="confirm-dialog__icon"
          :style="{
            background: `rgba(var(--v-theme-${confirmColor}), 0.12)`,
            color: `rgb(var(--v-theme-${confirmColor}))`,
          }"
          aria-hidden="true"
        >
          <v-icon :icon="icon" size="30" />
        </div>
        <div>
          <h2>{{ title }}</h2>
          <p>{{ text }}</p>
          <code v-if="subject" class="confirm-dialog__subject">{{
            subject
          }}</code>
        </div>
      </v-card-text>
      <v-card-actions class="px-7 pb-7 pt-0">
        <v-spacer />
        <v-btn variant="text" :disabled="loading" @click="model = false">
          Cancel
        </v-btn>
        <v-btn
          :color="confirmColor"
          variant="flat"
          :prepend-icon="confirmIcon"
          :loading="loading"
          @click="$emit('confirm')"
        >
          {{ confirmLabel }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<style scoped>
.confirm-dialog {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 18px;
}
.confirm-dialog__icon {
  display: grid;
  width: 54px;
  height: 54px;
  place-items: center;
  border-radius: 16px;
}
.confirm-dialog h2 {
  margin: 2px 0 8px;
  color: rgb(var(--gfw-text));
  font-size: 1.35rem;
}
.confirm-dialog p {
  margin: 0 0 14px;
  color: rgb(var(--gfw-text-secondary));
  line-height: 1.55;
}
.confirm-dialog__subject {
  display: inline-block;
  padding: 7px 10px;
  border: 1px solid rgba(var(--gfw-border), 0.65);
  border-radius: 9px;
  background: rgba(var(--gfw-muted), 0.7);
}
</style>
