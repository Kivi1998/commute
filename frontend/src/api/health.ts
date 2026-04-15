import { get } from './client'

export interface HealthResponse {
  status: string
  version: string
  uptime_seconds: number
  dependencies: Record<string, string>
}

export function fetchHealth() {
  return get<HealthResponse>('/health')
}
