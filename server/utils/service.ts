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
  latency: number | null
  logs: typeof schema.serviceLogs.$inferSelect[]
  uptime: number
  outages: { from: Temporal.ZonedDateTime, duration: Temporal.Duration }[]
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
        get uptime() {
          return this.upCount / this.total * 100
        },
        logs: [],
        get latency() {
          return this.logs.at(-1)?.latency ?? null
        },
        get outages() {
          return getOutages(this.logs)
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
  const halfDayAgo = now.subtract(Temporal.Duration.from({ hours: 12 }))
  const halfDayAgoDate = new Date(halfDayAgo.epochMilliseconds)

  // Filter logs from the last 3 hours
  const recentLogs = logs.filter((log) => {
    const logDate = log.timestamp instanceof Date ? log.timestamp : new Date(log.timestamp)
    return logDate >= halfDayAgoDate
  })

  // If no logs in the last 3 hours, return down
  if (recentLogs.length === 0) {
    return ServiceStatus.Down
  }

  // Count successful logs (status === 'up')
  const upCount = recentLogs.filter(log => log.status === 'up').length
  const uptime = (upCount / recentLogs.length) * 100

  // Calculate status based on uptime percentage
  if (uptime >= 99.5) {
    return ServiceStatus.Operational
  }
  else if (uptime > 98) {
    return ServiceStatus.Degraded
  }
  else {
    return ServiceStatus.Down
  }
}

export function getOutages(histories: typeof schema.serviceLogs.$inferSelect[]) {
  // Returns an array of outage periods: [{ from: Temporal.ZonedDateTime, duration: Temporal.Duration }]
  // Outage is defined as contiguous logs where status !== 'up'

  if (!histories || histories.length === 0)
    return []

  // Make sure logs are ordered
  const sorted = [...histories].sort((a, b) => {
    const aDate = a.timestamp instanceof Date ? a.timestamp.getTime() : new Date(a.timestamp).getTime()
    const bDate = b.timestamp instanceof Date ? b.timestamp.getTime() : new Date(b.timestamp).getTime()
    return aDate - bDate
  })

  const outages: { from: Temporal.ZonedDateTime, duration: Temporal.Duration }[] = []

  let inOutage = false
  let outageStart: Temporal.ZonedDateTime | null = null
  let lastTimestamp: Temporal.ZonedDateTime | null = null

  for (const log of sorted) {
    const logDate = log.timestamp instanceof Date
      ? Temporal.Instant.fromEpochMilliseconds(log.timestamp.getTime()).toZonedDateTimeISO('UTC')
      : Temporal.Instant.fromEpochMilliseconds(new Date(log.timestamp).getTime()).toZonedDateTimeISO('UTC')

    if (log.status !== 'up') {
      if (!inOutage) {
        inOutage = true
        outageStart = logDate
      }
    }
    else {
      if (inOutage && outageStart) {
        // Outage ends at previous log's timestamp
        const duration = logDate.since(outageStart)
        outages.push({ from: outageStart, duration })
        inOutage = false
        outageStart = null
      }
    }
    lastTimestamp = logDate
  }

  // If still in outage at the end, measure until lastTimestamp (last log)
  if (inOutage && outageStart && lastTimestamp) {
    const duration = lastTimestamp.since(outageStart)
    outages.push({ from: outageStart, duration })
  }

  return outages
}
