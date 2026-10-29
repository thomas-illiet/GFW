<script setup lang="ts">
import type { TrustedService } from '~/types/api'

const open = defineModel<boolean>({ required: true })
const props = defineProps<{
  loading: boolean
  service?: TrustedService | null
}>()
const emit = defineEmits<{
  save: [
    {
      name: string
      description: string
      isActive: boolean
      hostnames: string[]
    },
  ]
}>()
const form = reactive({
  name: '',
  description: '',
  hostnames: '',
  isActive: true,
})
const valid = shallowRef(false)
const required = (value: string) => Boolean(value?.trim()) || 'Required field'
const hostnamesValid = (value: string) => {
  const hostnames = value
    .split('\n')
    .map((item) => item.trim())
    .filter(Boolean)
  return (
    (hostnames.length > 0 &&
      hostnames.every((hostname) =>
        /^(?=.{1,253}$)(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$/i.test(
          hostname,
        ),
      )) ||
    'Enter at least one valid DNS hostname'
  )
}
watch(
  [open, () => props.service],
  ([isOpen]) => {
    if (!isOpen) return
    const service = props.service
    Object.assign(form, {
      name: service?.Name ?? '',
      description: service?.Description ?? '',
      hostnames:
        service?.Hostnames.map((host) => host.Hostname).join('\n') ?? '',
      isActive: service?.IsActive ?? true,
    })
  },
  { immediate: true },
)
function save() {
  if (!valid.value) return
  emit('save', {
    name: form.name,
    description: form.description,
    isActive: form.isActive,
    hostnames: form.hostnames
      .split('\n')
      .map((v) => v.trim())
      .filter(Boolean),
  })
}
</script>
<template>
  <v-dialog v-model="open" max-width="620"
    ><v-card
      :title="service ? 'Edit trusted service' : 'Create trusted service'"
      ><v-card-text
        ><v-form id="trusted" v-model="valid" @submit.prevent="save"
          ><v-text-field
            v-model="form.name"
            label="Service name"
            prepend-inner-icon="mdi-shield-check-outline"
            :rules="[required]"
            autofocus /><v-textarea
            v-model="form.description"
            label="Description"
            prepend-inner-icon="mdi-text-box-outline"
            rows="2" /><v-textarea
            v-model="form.hostnames"
            label="DNS hostnames"
            prepend-inner-icon="mdi-dns-outline"
            hint="One hostname per line"
            :rules="[required, hostnamesValid]"
            persistent-hint
            rows="4" /><v-switch
            v-model="form.isActive"
            label="Active" /></v-form></v-card-text
      ><v-card-actions
        ><v-spacer /><v-btn variant="text" @click="open = false">Cancel</v-btn
        ><v-btn
          color="primary"
          :prepend-icon="service ? 'mdi-content-save-outline' : 'mdi-plus'"
          type="submit"
          form="trusted"
          :loading="loading"
          :disabled="!valid"
          >{{ service ? 'Save changes' : 'Create service' }}</v-btn
        ></v-card-actions
      ></v-card
    ></v-dialog
  >
</template>
