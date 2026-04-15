import { del, get, post, put } from './client'

export interface HomeAddress {
  id: number
  user_id: number
  alias: string
  address: string
  province?: string
  city?: string
  district?: string
  longitude: number
  latitude: number
  is_default: boolean
  note?: string
  created_at: string
  updated_at: string
}

export interface HomeAddressCreateInput {
  alias: string
  address: string
  province?: string
  city?: string
  district?: string
  longitude: number
  latitude: number
  is_default?: boolean
  note?: string
}

export type HomeAddressUpdateInput = Partial<HomeAddressCreateInput>

export const listAddresses = () =>
  get<{ list: HomeAddress[] }>('/addresses').then((r) => r.list)

export const createAddress = (body: HomeAddressCreateInput) =>
  post<HomeAddress>('/addresses', body)

export const updateAddress = (id: number, body: HomeAddressUpdateInput) =>
  put<HomeAddress>(`/addresses/${id}`, body)

export const deleteAddress = (id: number) => del<{ id: number }>(`/addresses/${id}`)

export const setDefaultAddress = (id: number) =>
  post<HomeAddress>(`/addresses/${id}/set-default`)
