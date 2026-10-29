import type { Role, User } from '~/types/api'
const ranks: Record<Role, number> = { viewer: 1, editor: 2, admin: 3 }
export const useAuthStore = defineStore('auth', () => {
  const user = shallowRef<User | null>(null)
  const ready = shallowRef(false)
  const error = shallowRef<string | null>(null)
  async function initialize() {
    if (ready.value) return
    try {
      user.value = await $fetch<User>('/auth/session', {
        credentials: 'include',
      })
    } catch (cause) {
      user.value = null
      error.value =
        cause instanceof Error ? cause.message : 'Authentication failed'
    } finally {
      ready.value = true
    }
  }
  function can(role: Role) {
    return Boolean(user.value && ranks[user.value.role] >= ranks[role])
  }
  function login() {
    window.location.assign('/auth/login')
  }
  async function logout() {
    await $fetch('/auth/logout', { method: 'POST', credentials: 'include' })
    user.value = null
    await navigateTo('/login')
  }
  return {
    user: readonly(user),
    ready: readonly(ready),
    error: readonly(error),
    initialize,
    can,
    login,
    logout,
  }
})
