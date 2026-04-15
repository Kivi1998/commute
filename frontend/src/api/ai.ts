import { post } from './client'
import type { CompanyType } from './profile'

export interface AIRecommendInput {
  city: string
  position: string
  experience_years?: number
  company_types?: CompanyType[]
  count?: number
  force_refresh?: boolean
}

export interface AIRecommendedCompany {
  name: string
  category: string
  industry: string
  address_hint: string
  reason: string
  resolved_address?: string
  resolved_longitude?: number
  resolved_latitude?: number
  resolved_province?: string
  resolved_city?: string
  resolved_district?: string
  location_confident: boolean
}

export interface AIRecommendResult {
  from_cache: boolean
  cached_at?: string
  expires_at?: string
  companies: AIRecommendedCompany[]
  token_input?: number
  token_output?: number
}

export const recommendCompanies = (body: AIRecommendInput) =>
  post<AIRecommendResult>('/ai/recommend/companies', body)
