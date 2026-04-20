import { reactive } from 'vue'

const toastState = reactive({
  message: '',
  visible: false,
})

let toastTimer: number | undefined

const showToast = (message: string, duration = 2500) => {
  toastState.message = message
  toastState.visible = true
  if (toastTimer) {
    window.clearTimeout(toastTimer)
  }
  toastTimer = window.setTimeout(() => {
    toastState.visible = false
  }, duration)
}

export const useToast = () => ({
  toastState,
  showToast,
})

