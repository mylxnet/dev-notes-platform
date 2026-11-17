import { ref } from 'vue'

export const toastMessage = ref('')
let timer: number | undefined

export function showToast(message: string): void {
  toastMessage.value = message
  if (timer) window.clearTimeout(timer)
  timer = window.setTimeout(() => {
    toastMessage.value = ''
  }, 2400)
}
