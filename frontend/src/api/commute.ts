import { del, get, post } from './client'
import type { HomeAddress } from './address'

export type TransportMode = 'transit' | 'driving' | 'cycling' | 'walking'
export type CommuteDirection = 'to_work' | 'to_home'
export type TimeStrategy = 'depart_at' | 'arrive_by'

export interface CommuteTimeSpec {
  strategy?: TimeStrategy
  time: string // HH:MM
}

export interface CommuteCalculateInput {
  home_id: number
  company_ids: number[]
  transport_modes: TransportMode[]
  morning: CommuteTimeSpec
  evening: CommuteTimeSpec
  weekday?: number
  buffer_minutes?: number
  force_refresh?: boolean
  save_query?: boolean
}

export interface CommuteResultItem {
  direction: CommuteDirection
  transport_mode: TransportMode
  depart_time: string
  arrive_time: string
  duration_min: number
  duration_min_raw: number
  distance_km: number
  cost_yuan?: number
  transfer_count?: number
  polyline: string
  from_cache: boolean
  result_id: number
}

export interface CommuteCalcError {
  direction: CommuteDirection
  transport_mode: TransportMode
  message: string
}

export interface CompanyCommute {
  company_id: number
  company_name: string
  company_longitude: number
  company_latitude: number
  items: CommuteResultItem[]
  errors: CommuteCalcError[]
}

export interface CommuteSummary {
  total_companies: number
  total_calculations: number
  cache_hits: number
  failures: number
}

export interface CommuteCalculateResponse {
  query_id?: number
  home: HomeAddress
  weekday: number
  buffer_minutes: number
  results: CompanyCommute[]
  summary: CommuteSummary
}

export interface CommuteResultDetail extends CommuteResultItem {
  id: number
  query_id?: number
  home_id: number
  company_id: number
  weekday: number
  route_detail: unknown
  calculated_at: string
  expires_at: string
  is_failed: boolean
}

export const calculateCommute = (body: CommuteCalculateInput) =>
  post<CommuteCalculateResponse | { warning: string; result: CommuteCalculateResponse }>(
    '/commute/calculate',
    body,
  )

export const getCommuteResult = (id: number) =>
  get<CommuteResultDetail>(`/commute/results/${id}`)

// --- 历史查询 ---

export interface CommuteQueryListItem {
  id: number
  user_id: number
  home_id: number
  transport_modes: TransportMode[]
  morning_strategy: TimeStrategy
  morning_time: string
  evening_strategy: TimeStrategy
  evening_time: string
  weekday: number
  buffer_minutes: number
  created_at: string
  home_alias: string
  home_address: string
  company_count: number
  company_names: string[]
}

export interface CommuteQueryDetail {
  id: number
  user_id: number
  home_id: number
  transport_modes: TransportMode[]
  morning_strategy: TimeStrategy
  morning_time: string
  evening_strategy: TimeStrategy
  evening_time: string
  weekday: number
  buffer_minutes: number
  created_at: string
}

export const listCommuteQueries = () =>
  get<{ list: CommuteQueryListItem[] }>('/commute/queries').then((r) => r.list)

export const getCommuteQuery = (id: number) =>
  get<CommuteQueryDetail>(`/commute/queries/${id}`)

export const listResultsByQuery = (id: number) =>
  get<{ list: CommuteResultDetail[] }>(`/commute/queries/${id}/results`).then((r) => r.list)

export const deleteCommuteQuery = (id: number) =>
  del<{ id: number }>(`/commute/queries/${id}`)
