import { schema } from 'hub:db'

export async function checkService(service: typeof schema.services.$inferSelect) {
  const performance = new Performance()
  const startTime = performance.now()

  let status: 'up' | 'timeout' | 'error' = 'up'

  let errorMessage: string = ''

  const response = await $fetch.raw(service.url, {
    method: service.method || 'GET',
    headers: {
      'User-Agent': 'Pulse/1.0',
    },
    timeout: 30000,
    retry: 3,
    onRequestError({ error }) {
      status = 'error'
      errorMessage = error.message
    },
    onResponseError({ error }) {
      if (error?.name === 'TimeoutError') {
        status = 'timeout'
        errorMessage = 'Request timed out'
      }
      else {
        status = 'error'
        errorMessage = error?.message ?? 'Unknown error'
      }
    },
  })

  const statusCode = response.status

  const latency = performance.now() - startTime

  // Create log entry
  await db.insert(schema.serviceLogs).values({
    serviceId: service.id,

    status,
    latency,

    statusCode,
    errorMessage,
  })

  await kv.set(
    `service:${service.id}`,
    status === 'up' ? ServiceStatus.Up : ServiceStatus.Down,
  )
}
