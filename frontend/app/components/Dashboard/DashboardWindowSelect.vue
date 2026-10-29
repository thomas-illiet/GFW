<script setup lang="ts">
const model = defineModel<'24h' | '7d' | '30d'>({ required: true })
</script>
<template>
  <v-btn-toggle v-model="model" mandatory color="primary" variant="outlined"
    ><v-btn value="24h">24 hours</v-btn><v-btn value="7d">7 days</v-btn
    ><v-btn value="30d">30 days</v-btn></v-btn-toggle
  >
</template>
