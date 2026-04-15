import { get, put } from './client'

export type CompanyType = 'big_tech' | 'mid_tech' | 'startup' | 'foreign' | 'other'

export interface Profile {
  id: number
  user_id: number
  current_city: string
  current_city_code?: string
  target_position: string
  experience_years?: number
  preferred_company_types: CompanyType[]
  created_at: string
  updated_at: string
}

export interface ProfileUpsertInput {
  current_city: string
  current_city_code?: string
  target_position: string
  experience_years?: number
  preferred_company_types?: CompanyType[]
}

export const fetchProfile = () => get<Profile | null>('/profile')
export const upsertProfile = (body: ProfileUpsertInput) => put<Profile>('/profile', body)
