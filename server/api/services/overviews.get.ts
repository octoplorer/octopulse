import { and, eq, gte } from 'drizzle-orm'
import { db, schema } from 'hub:db'

interface ServiceLog {
  timestamp: Date
  latency: number | null
  status: 'up' | 'timeout' | 'error'
}

interface Overview {
  service: {
    id: number
    name: string
    tags: string[]
    url: string
    status: ServiceStatus
  }
  logs: ServiceLog[]
}

export default defineEventHandler(async () => {
  const services = await db.select().from(schema.services)
  const overviews: Overview[] = []

  const now = Temporal.Now.instant().toZonedDateTimeISO('UTC')
  const thirtyDaysAgo = now.subtract(Temporal.Duration.from({ days: 30 }))
  const thirtyDaysAgoDate = new Date(thirtyDaysAgo.toInstant().epochMilliseconds)

  for (const service of services) {
    const status = await kv.get<ServiceStatus>(`service:${service.id}`)
    const serviceLogs = await db
      .select()
      .from(schema.serviceLogs)
      .where(
        and(
          eq(schema.serviceLogs.serviceId, service.id),
          gte(schema.serviceLogs.timestamp, thirtyDaysAgoDate),
        ),
      )
      .orderBy(schema.serviceLogs.timestamp)

    const logs: ServiceLog[] = serviceLogs.map(log => ({
      status: log.status,
      latency: log.latency,
      timestamp: new Date(log.timestamp),
    }))

    overviews.push({
      logs,
      service: {
        id: service.id,
        name: service.name,
        tags: service.tags,
        url: service.url,
        status: status ?? ServiceStatus.Down,
      },
    })
  }
  return {
    overviews,
    timestamp: new Date(),
  }
})
