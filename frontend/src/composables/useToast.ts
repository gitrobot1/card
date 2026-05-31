import { ref } from 'vue'

export type ToastType = 'success' | 'error' | 'warning' | 'info'

export interface ToastItem {
  id: number
  message: string
  type: ToastType
}

const toasts = ref<ToastItem[]>([])
let toastSeq = 0

export function showToast(message: string, type: ToastType = 'info', duration = 3200) {
  const id = ++toastSeq
  toasts.value = [...toasts.value, { id, message, type }]
  window.setTimeout(() => {
    toasts.value = toasts.value.filter((item) => item.id !== id)
  }, duration)
}

export function useToastState() {
  return { toasts }
}
