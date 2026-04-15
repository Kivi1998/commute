import axios, { type AxiosError, type AxiosInstance } from 'axios'
import { message } from 'ant-design-vue'

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
  request_id?: string
}

export class ApiError extends Error {
  code: number
  bizMessage: string
  data: unknown
  requestId?: string

  constructor(code: number, bizMessage: string, data: unknown, requestId?: string) {
    super(bizMessage)
    this.name = 'ApiError'
    this.code = code
    this.bizMessage = bizMessage
    this.data = data
    this.requestId = requestId
  }
}

const http: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

http.interceptors.response.use(
  (res) => {
    const body = res.data as ApiResponse
    if (body && typeof body.code === 'number') {
      if (body.code === 0) return res
      const err = new ApiError(body.code, body.message, body.data, body.request_id)
      message.error(body.message || '请求失败')
      return Promise.reject(err)
    }
    return res
  },
  (err: AxiosError<ApiResponse>) => {
    const body = err.response?.data
    const apiErr = new ApiError(
      body?.code ?? -1,
      body?.message || err.message || '网络错误',
      body?.data,
      body?.request_id,
    )
    message.error(apiErr.bizMessage)
    return Promise.reject(apiErr)
  },
)

export async function get<T>(url: string, params?: Record<string, unknown>): Promise<T> {
  const res = await http.get<ApiResponse<T>>(url, { params })
  return res.data.data
}

export async function post<T>(url: string, body?: unknown): Promise<T> {
  const res = await http.post<ApiResponse<T>>(url, body)
  return res.data.data
}

export async function put<T>(url: string, body?: unknown): Promise<T> {
  const res = await http.put<ApiResponse<T>>(url, body)
  return res.data.data
}

export async function patch<T>(url: string, body?: unknown): Promise<T> {
  const res = await http.patch<ApiResponse<T>>(url, body)
  return res.data.data
}

export async function del<T>(url: string): Promise<T> {
  const res = await http.delete<ApiResponse<T>>(url)
  return res.data.data
}

export default http
