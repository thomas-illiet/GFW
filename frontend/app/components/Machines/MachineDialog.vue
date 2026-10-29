<script setup lang="ts">
import type { Ecosystem } from '~/types/api'
const open = defineModel<boolean>({ required: true })
defineProps<{ ecosystems: Ecosystem[]; loading: boolean }>()
const emit = defineEmits<{
  save: [{ hostname: string; ecosystemID: string; ipv4: string }]
}>()
const form = reactive({ hostname: '', ecosystemID: '', ipv4: '' })
const valid = shallowRef(false)
const required = (value: string) => Boolean(value?.trim()) || 'Required field'
const ipv4 = (value: string) => {
  const parts = value.split('.')
  return (
    (parts.length === 4 &&
      parts.every((part) => /^\d{1,3}$/.test(part) && Number(part) <= 255)) ||
    'Enter a valid IP address'
  )
}
function submit() {
  if (valid.value) emit('save', { ...form, hostname: form.hostname.trim() })
}
watch(open, (value) => {
  if (value)
    Object.assign(form, {
      hostname: '',
      ecosystemID: '',
      ipv4: '',
    })
})
</script>
<template>
  <v-dialog
    v-model="open"
    max-width="560"
    aria-labelledby="machine-dialog-title"
    ><v-card class="machine-dialog" rounded="xl"
      ><v-card-text class="pa-7"
        ><header class="machine-dialog__header">
          <div class="machine-dialog__icon" aria-hidden="true">
            <v-icon icon="mdi-server-plus-outline" size="28" />
          </div>
          <div>
            <p>Manual address</p>
            <h2 id="machine-dialog-title">Add machine</h2>
            <span>Add an address managed directly in Great Firewall.</span>
          </div>
        </header>
        <v-form id="machine-form" v-model="valid" @submit.prevent="submit"
          ><v-text-field
            v-model="form.hostname"
            label="Hostname"
            prepend-inner-icon="mdi-server-outline"
            :rules="[required]"
            autofocus
            hint="Use the machine name shown in your inventory."
            persistent-hint
            required /><v-select
            v-model="form.ecosystemID"
            :items="ecosystems"
            item-title="DisplayName"
            item-value="ID"
            label="Ecosystem"
            prepend-inner-icon="mdi-cube-outline"
            placeholder="Select an ecosystem"
            :rules="[required]"
            required /><v-text-field
            v-model="form.ipv4"
            label="IP address"
            placeholder="10.0.0.10"
            prepend-inner-icon="mdi-ip-network-outline"
            :rules="[required, ipv4]"
            hint="Enter one exact address, for example 10.0.0.10."
            persistent-hint
            required /></v-form></v-card-text
      ><v-card-actions class="px-7 pb-7 pt-0"
        ><v-spacer /><v-btn variant="text" @click="open = false">Cancel</v-btn
        ><v-btn
          color="primary"
          prepend-icon="mdi-content-save-outline"
          type="submit"
          form="machine-form"
          :loading="loading"
          :disabled="!valid || !ecosystems.length"
          >Save machine</v-btn
        ></v-card-actions
      ></v-card
    ></v-dialog
  >
</template>

<style scoped>
.machine-dialog {
  border: 1px solid rgba(var(--gfw-ember), 0.18);
}
.machine-dialog__header {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 16px;
  margin-bottom: 22px;
}
.machine-dialog__icon {
  display: grid;
  width: 54px;
  height: 54px;
  place-items: center;
  border-radius: 16px;
  background: rgba(var(--gfw-ember), 0.1);
  color: rgb(var(--gfw-ember));
}
.machine-dialog__header p {
  margin: 0 0 3px;
  color: rgb(var(--gfw-ember));
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}
.machine-dialog__header h2 {
  margin: 0;
  font-size: 1.45rem;
}
.machine-dialog__header span {
  display: block;
  margin-top: 5px;
  color: rgb(var(--gfw-text-secondary));
}
#machine-form {
  display: grid;
  gap: 10px;
}
@media (max-width: 600px) {
  .machine-dialog__header {
    grid-template-columns: 1fr;
  }
}
</style>
