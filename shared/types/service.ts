export enum ServiceStatus {
  /** Service is up */
  Up = 'UP',
  /** Service is down */
  Down = 'DOWN',
}

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
  timestamp: Date
}

export interface ServiceOverview {
  logs: Pick<ServiceLog, 'status' | 'latency' | 'timestamp'>[]
  service: Service
}
