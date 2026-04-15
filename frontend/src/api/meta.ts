import { get } from './client'

export interface EnumItem {
  value: string
  label: string
  icon?: string
}

export interface Enums {
  company_type: EnumItem[]
  company_status: EnumItem[]
  company_source: EnumItem[]
  transport_mode: EnumItem[]
  time_strategy: EnumItem[]
  commute_direction: EnumItem[]
}

let cache: Enums | null = null

export async function fetchEnums(): Promise<Enums> {
  if (cache) return cache
  cache = await get<Enums>('/meta/enums')
  return cache
}
