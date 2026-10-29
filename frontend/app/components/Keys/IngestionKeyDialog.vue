<script setup lang="ts">
const open = defineModel<boolean>({ required: true })
defineProps<{ loading: boolean }>()
defineEmits<{ save: [{ name: string }] }>()
const name = shallowRef('')
const valid = shallowRef(false)
const required = (value: string) => Boolean(value?.trim()) || 'Required field'
watch(open, (v) => {
  if (v) name.value = ''
})
</script>
<template>
  <v-dialog v-model="open" max-width="500"
    ><v-card title="Create ingestion key"
      ><v-card-text
        ><v-form
          id="key"
          v-model="valid"
          @submit.prevent="valid && $emit('save', { name: name.trim() })"
          ><v-text-field
            v-model="name"
            label="Key name"
            prepend-inner-icon="mdi-key-outline"
            hint="Identify the inventory source"
            :rules="[required]"
            autofocus
            persistent-hint /></v-form></v-card-text
      ><v-card-actions
        ><v-spacer /><v-btn variant="text" @click="open = false">Cancel</v-btn
        ><v-btn
          color="primary"
          prepend-icon="mdi-key-plus"
          type="submit"
          form="key"
          :loading="loading"
          :disabled="!valid"
          >Create key</v-btn
        ></v-card-actions
      ></v-card
    ></v-dialog
  >
</template>
