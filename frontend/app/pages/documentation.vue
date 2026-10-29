<script setup lang="ts">
definePageMeta({ role: 'viewer' })
</script>

<template>
  <v-container fluid class="gfw-page">
    <v-row>
      <v-col cols="12">
        <AppPageHeader
          title="API Documentation"
          copy="Explore the Great Firewall API contract and run authenticated requests."
          icon="mdi-api"
        />
      </v-col>
      <v-col cols="12">
        <v-card class="documentation-card pa-3 pa-md-5">
          <ClientOnly>
            <DocumentationApiSwagger />
            <template #fallback>
              <div class="documentation-loading">
                <v-progress-circular indeterminate color="primary" />
                <span>Loading API documentation…</span>
              </div>
            </template>
          </ClientOnly>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped>
.documentation-card {
  overflow: hidden;
}
.documentation-loading {
  display: flex;
  min-height: 360px;
  align-items: center;
  justify-content: center;
  gap: 14px;
  color: rgb(var(--gfw-text-secondary));
}
</style>
