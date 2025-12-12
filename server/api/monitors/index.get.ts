import { db, schema } from 'hub:db';

/**
 * GET /monitors
 * Get all monitoring details (authentication required)
 */
export default defineEventHandler(async (event) => {
  // Verify authentication
  verifyAuth(event);

  // Get all monitors with full details
  const allMonitors = await db.select().from(schema.monitors);

  return {
    monitors: allMonitors,
  };
});
