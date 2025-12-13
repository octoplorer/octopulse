import { db, schema } from 'hub:db'

/**
 * GET /services
 * Get all services (authentication required)
 */
export default defineEventHandler(async (event) => {
  // Verify authentication
  await requireUserSession(event)

  // Get all monitors with full details
  const services = await db.select().from(schema.services)

  return {
    services,
  }
})
