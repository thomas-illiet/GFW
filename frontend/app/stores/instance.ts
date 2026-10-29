import type {
  InstanceEnvironmentState,
  InstanceResponse,
} from '~/types/instance'
import { isInstanceEnvironment } from '~/types/instance'
import { localInstanceEnvironment } from '~/utils/instance'

export const useInstanceStore = defineStore('instance', () => {
  const environment = shallowRef<InstanceEnvironmentState>(
    import.meta.client
      ? localInstanceEnvironment(window.location.hostname)
      : 'unknown',
  )
  const ready = shallowRef(false)
  let initializationPromise: Promise<void> | null = null

  function initialize() {
    if (initializationPromise) return initializationPromise
    initializationPromise = $fetch<InstanceResponse>('/api/v1/public/instance')
      .then((response) => {
        if (isInstanceEnvironment(response.environment)) {
          environment.value = response.environment
        }
      })
      .finally(() => {
        ready.value = true
      })
      .catch(() => undefined)
    return initializationPromise
  }

  return {
    environment: readonly(environment),
    ready: readonly(ready),
    initialize,
  }
})
