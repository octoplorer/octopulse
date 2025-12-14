import { isNull } from 'drizzle-orm'
import { db, schema } from 'hub:db'

/**
 * GET /services
 * Get all services (authentication required)
 */
export default defineEventHandler(async (event) => {
  // Verify authentication
  await requireUserSession(event)

  // Get all monitors with full details (excluding soft-deleted)
  const services = await db
    .select()
    .from(schema.services)
    .where(isNull(schema.services.deletedAt))

  return {
    services,
  }
})
