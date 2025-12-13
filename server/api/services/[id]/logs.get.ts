import { eq } from 'drizzle-orm'
import { db, schema } from 'hub:db'

/**
 * GET /monitors/:id/logs
 * Get logs for a specific monitor (authentication required)
 */
export default defineEventHandler(async (event) => {
  // Verify authentication
  await requireUserSession(event)

  const id = getRouterParam(event, 'id')

  if (!id) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Bad Request',
      message: 'Missing monitor ID',
    })
  }

  const serviceId = Number.parseInt(id, 10)

  if (Number.isNaN(serviceId)) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Bad Request',
      message: 'Invalid service ID',
    })
  }

  // Check if monitor exists
  const existingService = await db.select().from(schema.services).where(eq(schema.services.id, serviceId)).limit(1)

  if (existingService.length === 0) {
    throw createError({
      statusCode: 404,
      statusMessage: 'Not Found',
      message: 'Monitor not found',
    })
  }

  // Get query parameters for pagination
  const query = getQuery(event)
  const limit = query.limit ? Number.parseInt(query.limit as string, 10) : 30
  const offset = query.offset ? Number.parseInt(query.offset as string, 10) : 0

  const logs = await getServiceLogs(serviceId, limit, offset)

  return logs
})
