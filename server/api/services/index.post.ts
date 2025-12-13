import { db, schema } from 'hub:db'
import { z } from 'zod'

const createServiceSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  url: z.url('Invalid URL format'),
  method: z.enum(['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS']).default('GET'),
  interval: z.number().int().positive('Interval must be a positive integer').default(60),
  tags: z.array(z.string()).default([]),
  type: z.enum(['public', 'private']).default('public'),
})

/**
 * POST /services
 * Add new service (authentication required)
 */
export default defineEventHandler(async (event) => {
  // Verify authentication
  verifyAuth(event)

  const validateBody = await readValidatedBody(event, createServiceSchema.safeParse)
  if (!validateBody.success) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Bad Request',
      message: validateBody.error.message,
    })
  }

  const body = validateBody.data

  // Validate required fields
  if (!body.name || !body.url) {
    throw createError({
      statusCode: 400,
      statusMessage: 'Bad Request',
      message: 'Missing required fields: name, url',
    })
  }

  // Insert new monitor
  const [newService] = await db.insert(schema.services).values({
    name: body.name,
    url: body.url,
    method: body.method || 'GET',
    interval: body.interval || 60, // Default 60 seconds
    tags: body.tags,
  }).returning()

  return {
    monitor: newService,
  }
})
