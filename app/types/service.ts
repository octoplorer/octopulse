import type { Incident } from './incident'

export interface Service {
  name: string
  tags: string[]
  url: string | null
  status: ServiceStatus
}

export interface ServiceLog {
  status: 'up' | 'timeout' | 'error'
  latency: number | null
  statusCode: number
  errorMessage: string | null
  timestamp: Temporal.PlainDateTime
}

export interface ServiceOverview {
  logs: {
    status: ServiceStatus
    /** Date */
    date: Temporal.PlainDate
    /** Uptime percentage */
    uptime: number
    /** Incidents */
    incidents: Incident[]
  }[]
  service: Service
}
