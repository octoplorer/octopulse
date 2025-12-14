import { and, eq, isNull } from 'drizzle-orm'
import { db, schema } from 'hub:db'

/**
 * DELETE /monitors/:id
 * Delete monitor (authentication required)
 * Soft delete: sets deletedAt timestamp instead of removing the record
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

  // Check if monitor exists and is not deleted
  const existingService = await db
    .select()
    .from(schema.services)
    .where(
      and(
        eq(schema.services.id, serviceId),
        isNull(schema.services.deletedAt),
      ),
    )
    .limit(1)

  if (existingService.length === 0) {
    throw createError({
      statusCode: 404,
      statusMessage: 'Not Found',
      message: 'Service not found',
    })
  }

  // Soft delete service: set deletedAt timestamp
  await db
    .update(schema.services)
    .set({ deletedAt: new Date() })
    .where(eq(schema.services.id, serviceId))

  return {
    success: true,
    message: 'Monitor deleted successfully',
  }
})
