import { db, schema } from 'hub:db'
import { checkService } from '~~/server/utils/service'

export default defineTask({
  meta: {
    name: 'check',
    description: 'Check status for all services',
  },
  async run() {
    // Fetch all services from database
    const services = await db.select().from(schema.services)

    // Process each service
    for (const service of services) {
      await checkService(service)
    }

    return {
      result: 'success',
    }
  },
})
