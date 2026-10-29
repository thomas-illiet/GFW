export default defineNuxtRouteMiddleware(async (to) => {
  const auth = useAuthStore()
  await auth.initialize()
  if (to.meta.auth === false) {
    if (to.path === '/login' && auth.user) return navigateTo('/')
    return
  }
  if (!auth.user)
    return navigateTo({ path: '/login', query: { redirect: to.fullPath } })
  const role = to.meta.role as 'viewer' | 'editor' | 'admin' | undefined
  if (role && !auth.can(role)) return navigateTo('/')
})
