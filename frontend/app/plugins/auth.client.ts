export default defineNuxtPlugin(async () => {
  await useAuthStore().initialize()
})
