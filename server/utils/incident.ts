import { and, eq } from 'drizzle-orm'
import { db, schema } from 'hub:db'
import { kv } from 'hub:kv'
import { ServiceStatus } from '~~/shared/types/service'

export async function checkIncident(service: typeof schema.services.$inferSelect) {
  const serviceStatus = await kv.get<ServiceStatus>(`service:${service.id}`)

  // Only create incident if service is down
  if (serviceStatus !== ServiceStatus.Down) {
    return null
  }

  // Check if there's already an ongoing incident for this service
  const existingIncidents = await db
    .select()
    .from(schema.incidents)
    .where(
      and(
        eq(schema.incidents.serviceId, service.id),
        eq(schema.incidents.status, 'ongoing'),
      ),
    )
    .limit(1)

  // If there's already an ongoing incident, don't create a new one
  if (existingIncidents.length > 0) {
    return existingIncidents[0]
  }

  // Create new incident
  const [incident] = await db.insert(schema.incidents).values({
    serviceId: service.id,
    startTime: new Date().toISOString(),
    status: 'ongoing',
    title: `${service.name} is down`,
    description: `Service ${service.name} (${service.url}) is currently experiencing issues.`,
  }).returning()

  return incident
}
