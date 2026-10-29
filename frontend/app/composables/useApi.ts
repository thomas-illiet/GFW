export function useApi() {
  const loading = shallowRef(false)
  const error = shallowRef<string | null>(null)
  async function request<T>(
    path: string,
    options: Parameters<typeof $fetch<T>>[1] = {},
  ) {
    loading.value = true
    error.value = null
    try {
      return await $fetch<T>(path, { credentials: 'include', ...options })
    } catch (cause) {
      error.value =
        cause instanceof Error ? cause.message : 'Unexpected API error'
      throw cause
    } finally {
      loading.value = false
    }
  }
  return { loading: readonly(loading), error: readonly(error), request }
}
