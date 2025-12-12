import { eq } from 'drizzle-orm'
import { db, schema } from 'hub:db'

/**
 * Task to check HTTP status codes for all monitors
 * Runs every minute via cron schedule
 */
export default defineTask({
  meta: {
    name: 'check-monitors',
    description: 'Check HTTP status codes for all websites',
  },
  async run() {
    // Fetch all monitors from database
    const allMonitors = await db.select().from(schema.monitors)

    // Process each monitor
    for (const monitor of allMonitors) {
      const startTime = Date.now()
      let statusCode: number | null = null
      let isFail = false
      let reason: string | null = null
      let responseBody: string | null = null

      try {
        // Prepare fetch options
        const fetchOptions: RequestInit = {
          method: monitor.method || 'GET',
          headers: {},
        }

        // Add user agent if specified
        if (monitor.userAgent) {
          fetchOptions.headers = {
            ...fetchOptions.headers,
            'User-Agent': monitor.userAgent,
          }
        }

        // Make HTTP request with timeout
        const controller = new AbortController()
        const timeoutId = setTimeout(() => controller.abort(), 30000) // 30 second timeout

        const response = await fetch(monitor.url, {
          ...fetchOptions,
          signal: controller.signal,
        })

        clearTimeout(timeoutId)

        statusCode = response.status
        responseBody = await response.text()

        // Check if status code indicates failure (non-2xx)
        if (statusCode < 200 || statusCode >= 300) {
          isFail = true
          reason = `HTTP ${statusCode}`
        }

        // Check for keyword if specified
        if (monitor.keyword && !isFail) {
          if (!responseBody.includes(monitor.keyword)) {
            isFail = true
            reason = `Keyword "${monitor.keyword}" not found in response`
          }
        }
      }
      catch (error) {
        // Handle network errors, timeouts, etc.
        isFail = true
        if (error instanceof Error) {
          if (error.name === 'AbortError') {
            reason = 'Request timeout (30s)'
          }
          else {
            reason = error.message || 'Network error'
          }
        }
        else {
          reason = 'Unknown error'
        }
      }

      const latency = Date.now() - startTime
      const now = new Date().toISOString()

      // Determine new status
      let newStatus: string
      let newRetryCount: number

      if (isFail) {
        // If monitor was UP, set to DOWN and reset retry count
        // If monitor was already DOWN, increment retry count
        if (monitor.status === 'UP') {
          newStatus = 'DOWN'
          newRetryCount = 1
        }
        else {
          newStatus = 'DOWN'
          newRetryCount = monitor.retryCount + 1
        }
      }
      else {
        // Success - set to UP and reset retry count
        newStatus = 'UP'
        newRetryCount = 0
      }

      // Update monitor status
      await db
        .update(schema.monitors)
        .set({
          status: newStatus,
          lastCheck: now,
          retryCount: newRetryCount,
        })
        .where(eq(schema.monitors.id, monitor.id))

      // Create log entry
      await db.insert(schema.logs).values({
        monitorId: monitor.id,
        statusCode,
        latency,
        isFail: isFail ? 1 : 0,
        reason,
        createdAt: now,
      })
    }

    return {
      result: 'success',
    }
  },
})
