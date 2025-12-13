/**
 * GET /services/public
 * Get public services data (no authentication required)
 */
export default defineEventHandler(async () => {
  throw createError({
    statusCode: 403,
    statusMessage: 'Forbidden',
    message: 'This endpoint is not implemented yet',
  })
})
