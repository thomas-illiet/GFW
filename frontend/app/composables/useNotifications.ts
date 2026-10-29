type NotificationTone = 'success' | 'info' | 'warning' | 'error'

interface AppNotification {
  message: string
  tone: NotificationTone
}

export function useNotifications() {
  const notification = useState<AppNotification | null>(
    'app-notification',
    () => null,
  )

  function show(message: string, tone: NotificationTone = 'success') {
    notification.value = { message, tone }
  }

  function clear() {
    notification.value = null
  }

  return {
    notification: readonly(notification),
    show,
    success: (message: string) => show(message, 'success'),
    clear,
  }
}
