import { useAuth } from '../composables/useAuth'
import { useToast } from '../composables/useToast'

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? ''

export class ApiError extends Error {
  code?: number
  constructor(message: string, code?: number) {
    super(message)
    this.name = 'ApiError'
    this.code = code
  }
}

type ApiResponse<T> = {
  code: number
  message: string
  data: T
}

const request = async <T>(path: string, options: RequestInit = {}) => {
  const { token, clearToken } = useAuth()
  const { showToast } = useToast()

  const headers: Record<string, string> = options.headers
    ? (options.headers as Record<string, string>)
    : {}

  if (!headers['Content-Type'] && options.body) {
    headers['Content-Type'] = 'application/json'
  }

  if (token.value) {
    headers.Authorization = `Bearer ${token.value}`
  }

  let response: Response
  try {
    response = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers,
    })
  } catch (error) {
    showToast('网络异常，请稍后重试')
    throw new ApiError('网络异常，请稍后重试')
  }

  let payload: ApiResponse<T> | null = null
  try {
    payload = (await response.json()) as ApiResponse<T>
  } catch (error) {
    payload = null
  }

  if (!response.ok) {
    const message = payload?.message || '服务暂时不可用'
    showToast(message)
    if (payload && (payload.code === 4010 || payload.code === 4011)) {
      clearToken()
      window.location.href = '/login'
    }
    throw new ApiError(message, payload?.code)
  }

  if (!payload) {
    showToast('服务响应异常')
    throw new ApiError('服务响应异常')
  }

  if (payload.code !== 200) {
    showToast(payload.message || '请求失败')
    if (payload.code === 4010 || payload.code === 4011) {
      clearToken()
      window.location.href = '/login'
    }
    throw new ApiError(payload.message || '请求失败', payload.code)
  }

  return payload.data
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    }),
  put: <T>(path: string, body?: unknown) =>
    request<T>(path, {
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    }),
}
