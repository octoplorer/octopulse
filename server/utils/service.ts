import { desc, eq } from 'drizzle-orm'
import { db, schema } from 'hub:db'

export async function checkService(service: typeof schema.services.$inferSelect) {
  const startTime = Date.now()

  let status: 'up' | 'timeout' | 'error' = 'up'

  let errorMessage: string = ''

  const response = await $fetch.raw(service.url, {
    method: service.method || 'GET',
    headers: {
      'User-Agent': 'Pulse/1.0',
    },
    timeout: 30000,
    retry: 3,
    onRequestError({ error }) {
      status = 'error'
      errorMessage = error.message
    },
    onResponseError({ error }) {
      if (error?.name === 'TimeoutError') {
        status = 'timeout'
        errorMessage = 'Request timed out'
      }
      else {
        status = 'error'
        errorMessage = error?.message ?? 'Unknown error'
      }
    },
  })

  const statusCode = response.status

  const latency = Date.now() - startTime

  // Create log entry
  await db.insert(schema.serviceLogs).values({
    serviceId: service.id,

    status,
    latency,

    statusCode,
    errorMessage,
  })
}

export async function getServiceLogs(
  serviceId: number,
  limit: number = 30,
  offset: number = 0,
) {
  return db
    .select()
    .from(schema.serviceLogs)
    .where(eq(schema.serviceLogs.serviceId, serviceId))
    .orderBy(desc(schema.serviceLogs.timestamp))
    .limit(limit)
    .offset(offset)
}

export interface ServiceLogsByDays {
  total: number
  upCount: number
  logs: typeof schema.serviceLogs.$inferSelect[]
  uptime: number
}

export function manageServiceLogsByDays(
  logs: typeof schema.serviceLogs.$inferSelect[],
) {
  const now = Temporal.Now.instant().toZonedDateTimeISO('UTC')

  const serviceLogsByDays = new Map(
    Array.from({ length: 30 }, (_, i) => {
      const date = now.subtract(Temporal.Duration.from({ days: i }))
        .startOfDay()
        .toString()
      const logsByDay: ServiceLogsByDays = {
        get total() {
          return this.logs.length
        },
        get upCount() {
          return this.logs.filter(log => log.status === 'up').length
        },
        logs: [],
        get uptime() {
          return this.upCount / this.total * 100
        },
      }
      return [date, logsByDay]
    }),
  )

  for (const log of logs) {
    const date = log.timestamp.toTemporalInstant().toZonedDateTimeISO('UTC').startOfDay()
    const serviceLogsByDay = serviceLogsByDays.get(date.toString())
    if (serviceLogsByDay) {
      serviceLogsByDay.logs.push(log)
    }
  }

  return serviceLogsByDays
}

export function getServiceStatus(
  logs: typeof schema.serviceLogs.$inferSelect[],
): ServiceStatus {
  const now = Temporal.Now.instant()
  const threeHoursAgo = now.subtract(Temporal.Duration.from({ hours: 3 }))
  const threeHoursAgoDate = new Date(threeHoursAgo.epochMilliseconds)

  // Filter logs from the last 3 hours
  const recentLogs = logs.filter((log) => {
    const logDate = log.timestamp instanceof Date ? log.timestamp : new Date(log.timestamp)
    return logDate >= threeHoursAgoDate
  })

  // If no logs in the last 3 hours, return down
  if (recentLogs.length === 0) {
    return ServiceStatus.Down
  }

  // Count successful logs (status === 'up')
  const upCount = recentLogs.filter(log => log.status === 'up').length
  const uptime = (upCount / recentLogs.length) * 100

  // Calculate status based on uptime percentage
  if (uptime === 100) {
    return ServiceStatus.Operational
  }
  else if (uptime > 98) {
    return ServiceStatus.Degraded
  }
  else {
    return ServiceStatus.Down
  }
}
