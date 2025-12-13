import type { ServiceOverview } from '~~/shared/types/service'
import { db, schema } from 'hub:db'

export default defineEventHandler(async () => {
  const services = await db.select().from(schema.services)
  const overviews: ServiceOverview[] = []

  for (const service of services) {
    const status = await kv.get<ServiceStatus>(`service:${service.id}`)
    const serviceLogs = await getServiceLogs(service.id)

    const logs = serviceLogs.map(log => ({
      status: log.status,
      latency: log.latency,
      timestamp: new Date(log.timestamp),
    }))

    overviews.push({
      logs,
      service: {
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
