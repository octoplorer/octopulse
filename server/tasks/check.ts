import { isNull } from 'drizzle-orm'
import { db, schema } from 'hub:db'
import { checkService } from '~~/server/utils/service'

export default defineTask({
  meta: {
    name: 'check',
    description: 'Check status for all services',
  },
  async run() {
    // Fetch all services from database (excluding soft-deleted)
    const services = await db
      .select()
      .from(schema.services)
      .where(isNull(schema.services.deletedAt))

    // Process each service
    for (const service of services) {
      await checkService(service)
    }

    return {
      result: 'success',
    }
  },
})
