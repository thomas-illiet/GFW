<script setup lang="ts">
import SwaggerUIBundle from 'swagger-ui-dist/swagger-ui-es-bundle.js'
import type { SwaggerUIInstance } from 'swagger-ui-dist/swagger-ui-es-bundle.js'

let swagger: SwaggerUIInstance | undefined

onMounted(() => {
  swagger = SwaggerUIBundle({
    dom_id: '#gfw-swagger-ui',
    url: '/openapi.yaml',
    deepLinking: true,
    displayRequestDuration: true,
    docExpansion: 'list',
    tagsSorter: 'alpha',
    tryItOutEnabled: true,
    validatorUrl: null,
    withCredentials: true,
  })
})

onBeforeUnmount(() => swagger?.destroy?.())
</script>

<template>
  <div id="gfw-swagger-ui" class="swagger-host" />
</template>

<style src="swagger-ui-dist/swagger-ui.css"></style>
<style scoped>
.swagger-host {
  min-height: 560px;
}
.swagger-host :deep(.information-container),
.swagger-host :deep(.scheme-container),
.swagger-host :deep(.models),
.swagger-host :deep(.authorization__btn),
.swagger-host :deep(.download-url-wrapper) {
  display: none !important;
}
.swagger-host :deep(.swagger-ui) {
  padding-top: 0;
}
</style>
