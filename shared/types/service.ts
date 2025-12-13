export enum ServiceStatus {
  /** Service is operational */
  Operational = 'operational',
  /** Service is degraded */
  Degraded = 'degraded',
  /** Service is down */
  Down = 'down',
  /** Service is under maintenance */
  Maintenance = 'maintenance',
}

export interface Service {
  id: number
  name: string
  tags: string[]
  url: string | null
  status: ServiceStatus
}

export interface ServiceHistroy {
  uptime: number | null
  /** latest latency */
  latency: number | null
  /** date by day */
  date: Temporal.ZonedDateTime
}

export interface ServiceOverview {
  service: Service
  histories: ServiceHistroy[]
}
