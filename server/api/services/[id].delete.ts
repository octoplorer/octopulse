import { eq } from 'drizzle-orm'
import { db, schema } from 'hub:db'

/**
 * DELETE /monitors/:id
 * Delete monitor (authentication required)
 * Related logs will be deleted automatically due to cascade delete
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
      message: 'Service not found',
    })
  }

  // Delete service (logs will be deleted automatically due to cascade)
  await db.delete(schema.services).where(eq(schema.services.id, serviceId))

  return {
    success: true,
    message: 'Monitor deleted successfully',
  }
})
