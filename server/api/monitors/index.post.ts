import { db, schema } from 'hub:db'

/**
 * POST /monitors
 * Add new monitor (authentication required)
 */
export default defineEventHandler(async (event) => {
  // Verify authentication
  verifyAuth(event)

  const body = await readBody(event)

  // Validate required fields
  if (!body.name || !body.url) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Bad Request',
      message: 'Missing required fields: name, url',
    })
  }

  // Insert new monitor
  const newMonitor = await db.insert(schema.monitors).values({
    name: body.name,
    url: body.url,
    method: body.method || 'GET',
    interval: body.interval || 60, // Default 60 seconds
    status: 'UP', // Initial status
    retryCount: 0,
    keyword: body.keyword || null,
    userAgent: body.userAgent || null,
    createdAt: new Date().toISOString(),
  }).returning()

  return {
    monitor: newMonitor[0],
  }
})
