import { db ,schema} from 'hub:db';   
import { eq } from 'drizzle-orm';

/**
 * DELETE /monitors/:id
 * Delete monitor (authentication required)
 * Related logs will be deleted automatically due to cascade delete
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

  // Delete monitor (logs will be deleted automatically due to cascade)
  await db.delete(schema.monitors).where(eq(schema.monitors.id, monitorId));

  return {
    success: true,
    message: 'Monitor deleted successfully',
  };
});
