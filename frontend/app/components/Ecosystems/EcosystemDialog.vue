<script setup lang="ts">
const open = defineModel<boolean>({ required: true })
defineProps<{ loading: boolean }>()
defineEmits<{ save: [{ name: string; displayName: string }] }>()
const form = reactive({ name: '', displayName: '' })
const valid = shallowRef(false)
const required = (value: string) => Boolean(value?.trim()) || 'Required field'
const identifier = (value: string) =>
  /^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(value) ||
  'Use lowercase letters, numbers and single hyphens'
watch(open, (v) => {
  if (v) Object.assign(form, { name: '', displayName: '' })
})
</script>
<template>
  <v-dialog v-model="open" max-width="520"
    ><v-card title="Create ecosystem"
      ><v-card-text
        ><v-form
          id="eco"
          v-model="valid"
          @submit.prevent="
            valid &&
            $emit('save', {
              name: form.name.trim(),
              displayName: form.displayName.trim(),
            })
          "
          ><v-text-field
            v-model="form.name"
            label="Identifier"
            prepend-inner-icon="mdi-identifier"
            hint="Lowercase letters, numbers and hyphens"
            :rules="[required, identifier]"
            autofocus
            persistent-hint /><v-text-field
            v-model="form.displayName"
            label="Display name"
            prepend-inner-icon="mdi-label-outline"
            :rules="[required]" /></v-form></v-card-text
      ><v-card-actions
        ><v-spacer /><v-btn variant="text" @click="open = false">Cancel</v-btn
        ><v-btn
          color="primary"
          prepend-icon="mdi-plus"
          form="eco"
          type="submit"
          :loading="loading"
          :disabled="!valid"
          >Create</v-btn
        ></v-card-actions
      ></v-card
    ></v-dialog
  >
</template>
