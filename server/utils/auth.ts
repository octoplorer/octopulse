/* eslint-disable node/prefer-global/process */
import type { H3Event } from 'h3'

/**
 * Verify Bearer token authentication
 * Format: "Bearer PASSWORD"
 */
export function verifyAuth(event: H3Event): boolean {
  const authHeader = getHeader(event, 'authorization')

  if (!authHeader) {
    throw createError({
      statusCode: 401,
      statusMessage: 'Unauthorized',
      message: 'Missing Authorization header',
    })
  }

  const [authType, password] = authHeader.split(' ')

  if (authType !== 'Bearer' || !password) {
    throw createError({
      statusCode: 401,
      statusMessage: 'Unauthorized',
      message: 'Invalid Authorization header format. Expected: Bearer PASSWORD',
    })
  }

  // Get password from environment variable
  const validPassword = process.env.ADMIN_PASSWORD || 'admin'

  if (password !== validPassword) {
    throw createError({
      statusCode: 403,
      statusMessage: 'Forbidden',
      message: 'Invalid password',
    })
  }

  return true
}
