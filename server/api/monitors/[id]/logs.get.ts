import { db ,schema} from 'hub:db';  
import { eq, desc } from 'drizzle-orm';

/**
 * GET /monitors/:id/logs
 * Get logs for a specific monitor (authentication required)
 */
export default defineEventHandler(async (event) => {
  // Verify authentication
  verifyAuth(event);

  const id = getRouterParam(event, 'id');
  
  if (!id) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Bad Request',
      message: 'Missing monitor ID',
    });
  }

  const monitorId = parseInt(id, 10);
  
  if (isNaN(monitorId)) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Bad Request',
      message: 'Invalid monitor ID',
    });
  }
 
  // Check if monitor exists
  const existingMonitor = await db.select().from(schema.monitors).where(eq(schema.monitors.id, monitorId)).limit(1);
  
  if (existingMonitor.length === 0) {
    throw createError({
      statusCode: 404,
      statusMessage: 'Not Found',
      message: 'Monitor not found',
    });
  }

  // Get query parameters for pagination
  const query = getQuery(event);
  const limit = query.limit ? parseInt(query.limit as string, 10) : 100;
  const offset = query.offset ? parseInt(query.offset as string, 10) : 0;

  // Get logs for this monitor, ordered by created_at descending
  const monitorLogs = await db
    .select()
    .from(schema.logs)
    .where(eq(schema.logs.monitorId, monitorId))
    .orderBy(desc(schema.logs.createdAt))
    .limit(limit)
    .offset(offset);

  return {
    monitorId,
    logs: monitorLogs,
    total: monitorLogs.length,
  };
});
