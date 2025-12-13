import { and, desc, eq, gte } from 'drizzle-orm'
import { db, schema } from 'hub:db'

export default defineEventHandler(async () => {
  const services = await db.select().from(schema.services)
  const overviews: ServiceOverview[] = []

  const now = Temporal.Now.instant().toZonedDateTimeISO('UTC')
  const thirtyDaysAgo = now.subtract(Temporal.Duration.from({ days: 30 }))
  const thirtyDaysAgoDate = new Date(thirtyDaysAgo.toInstant().epochMilliseconds)

  for (const service of services) {
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

    // Get the latest log for last check time
    const latestLog = await db
      .select()
      .from(schema.serviceLogs)
      .where(eq(schema.serviceLogs.serviceId, service.id))
      .orderBy(desc(schema.serviceLogs.timestamp))
      .limit(1)
      .then(logs => logs[0] ?? null)

    const serviceLogsByDays = manageServiceLogsByDays(serviceLogs)
    const status = getServiceStatus(serviceLogs)

    const logs: ServiceHistroy[] = Array.from(serviceLogsByDays.entries())
      .map(([date, logsByDay]) => ({
        date: Temporal.ZonedDateTime.from(date),
        uptime: logsByDay.uptime,
        latency: logsByDay.latency,
        outages: logsByDay.outages,
      }))
      .sort((a, b) => a.date.epochMilliseconds - b.date.epochMilliseconds)

    // Convert latest log timestamp to Temporal.ZonedDateTime
    const lastCheckTime = latestLog
      ? Temporal.Instant.fromEpochMilliseconds(
          latestLog.timestamp instanceof Date
            ? latestLog.timestamp.getTime()
            : new Date(latestLog.timestamp).getTime(),
        ).toZonedDateTimeISO('UTC')
      : null

    overviews.push({
      histories: logs,
      service: {
        id: service.id,
        name: service.name,
        tags: service.tags,
        url: service.url,
        status: status ?? ServiceStatus.Down,
      },
      lastCheckTime,
    })
  }
  return {
    overviews,
    timestamp: Temporal.Now.instant().toString(),
  }
})
