import { get, post } from './client'

export interface User {
  id: number
  name?: string
  email?: string
  phone?: string
  created_at: string
}

export interface LoginInput {
  email: string
  password: string
}

export interface LoginResponse {
  token: string
  expires_at: string
  user: User
}

export const login = (body: LoginInput) => post<LoginResponse>('/auth/login', body)
export const fetchMe = () => get<User>('/auth/me')
