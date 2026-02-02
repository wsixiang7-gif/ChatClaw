import { ref, computed } from 'vue'

export interface ToastProps {
  id: string
  title?: string
  description?: string
  variant?: 'default' | 'success' | 'error'
  duration?: number
}

const toasts = ref<ToastProps[]>([])
const TOAST_LIMIT = 5
const TOAST_REMOVE_DELAY = 300

let count = 0

function genId() {
  count = (count + 1) % Number.MAX_VALUE
  return count.toString()
}

function addToast(props: Omit<ToastProps, 'id'>) {
  const id = genId()
  const newToast: ToastProps = {
    id,
    variant: 'default',
    duration: 3000,
    ...props,
  }

  toasts.value = [newToast, ...toasts.value].slice(0, TOAST_LIMIT)

  // 自动移除
  if (newToast.duration && newToast.duration > 0) {
    setTimeout(() => {
      dismissToast(id)
    }, newToast.duration)
  }

  return id
}

function dismissToast(id: string) {
  toasts.value = toasts.value.filter((t) => t.id !== id)
}

export function useToast() {
  return {
    toasts: computed(() => toasts.value),
    toast: addToast,
    dismiss: dismissToast,
  }
}

// 便捷方法
export const toast = {
  success: (message: string) => addToast({ title: message, variant: 'success' }),
  error: (message: string) => addToast({ title: message, variant: 'error' }),
  default: (message: string) => addToast({ title: message, variant: 'default' }),
}
