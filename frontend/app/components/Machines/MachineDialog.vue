<script setup lang="ts">
import type { Ecosystem } from '~/types/api'
const open = defineModel<boolean>({ required: true })
const props = defineProps<{ ecosystems: Ecosystem[]; loading: boolean }>()
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
      ecosystemID: props.ecosystems[0]?.ID ?? '',
      ipv4: '',
    })
})
</script>
<template>
  <v-dialog v-model="open" max-width="560"
    ><v-card title="Add machine"
      ><v-card-text
        ><v-form id="machine-form" v-model="valid" @submit.prevent="submit"
          ><v-text-field
            v-model="form.hostname"
            label="Hostname"
            prepend-inner-icon="mdi-server-outline"
            :rules="[required]"
            autofocus
            required /><v-select
            v-model="form.ecosystemID"
            :items="ecosystems"
            item-title="DisplayName"
            item-value="ID"
            label="Ecosystem"
            prepend-inner-icon="mdi-cube-outline"
            :rules="[required]"
            required /><v-text-field
            v-model="form.ipv4"
            label="IP address"
            placeholder="10.0.0.10"
            prepend-inner-icon="mdi-ip-network-outline"
            :rules="[required, ipv4]"
            required /></v-form></v-card-text
      ><v-card-actions
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
