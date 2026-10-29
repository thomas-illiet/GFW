<script setup lang="ts">
import type { DeepReadonly } from 'vue'

import type { Dashboard } from '~/types/api'

defineProps<{ data: DeepReadonly<Dashboard> }>()
</script>
<template>
  <v-row
    ><v-col cols="12" lg="8"
      ><DashboardEcosystemTable :items="data.ecosystems" /></v-col
    ><v-col cols="12" lg="4"
      ><DashboardTrafficMix
        :formats="data.formats"
        :sources="data.sources" /></v-col
  ></v-row>
</template>
