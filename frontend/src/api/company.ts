import type { CompanyType } from './profile'
import { del, get, patch, post, put } from './client'

export type CompanyStatus =
  | 'watching'
  | 'applied'
  | 'interviewing'
  | 'offered'
  | 'rejected'
  | 'archived'

export type CompanySource = 'ai_recommend' | 'manual'

export interface Company {
  id: number
  user_id: number
  name: string
  address: string
  province?: string
  city?: string
  district?: string
  longitude: number
  latitude: number
  category?: CompanyType
  industry?: string
  status: CompanyStatus
  source: CompanySource
  ai_reason?: string
  note?: string
  created_at: string
  updated_at: string
}

export interface CompanyCreateInput {
  name: string
  address: string
  province?: string
  city?: string
  district?: string
  longitude: number
  latitude: number
  category?: CompanyType
  industry?: string
  status?: CompanyStatus
  source?: CompanySource
  ai_reason?: string
  note?: string
}

export type CompanyUpdateInput = Partial<CompanyCreateInput>

export interface CompanyListQuery {
  status?: CompanyStatus
  category?: CompanyType
  keyword?: string
  page?: number
  page_size?: number
}

export interface CompanyListResult {
  list: Company[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export const listCompanies = (q: CompanyListQuery = {}) =>
  get<CompanyListResult>('/companies', q as Record<string, unknown>)

export const getCompany = (id: number) => get<Company>(`/companies/${id}`)

export const createCompany = (body: CompanyCreateInput) => post<Company>('/companies', body)

export const updateCompany = (id: number, body: CompanyUpdateInput) =>
  put<Company>(`/companies/${id}`, body)

export const updateCompanyStatus = (id: number, status: CompanyStatus) =>
  patch<Company>(`/companies/${id}/status`, { status })

export const deleteCompany = (id: number) => del<{ id: number }>(`/companies/${id}`)

export const batchCreateCompanies = (companies: CompanyCreateInput[]) =>
  post<{
    created: Company[]
    skipped: { name: string; reason: string }[]
    warning?: string
  }>('/companies/batch', { companies })
