import { computed, ref } from 'vue'

const token = ref<string>(localStorage.getItem('token') ?? '')
const subscribed = ref<boolean>(localStorage.getItem('subscribed') === 'true')

const isAuthed = computed(() => Boolean(token.value))

const setToken = (value: string) => {
  token.value = value
  localStorage.setItem('token', value)
}

const clearToken = () => {
  token.value = ''
  localStorage.removeItem('token')
}

const setSubscribed = (value: boolean) => {
  subscribed.value = value
  localStorage.setItem('subscribed', String(value))
}

export const useAuth = () => ({
  token,
  subscribed,
  isAuthed,
  setToken,
  clearToken,
  setSubscribed,
})

