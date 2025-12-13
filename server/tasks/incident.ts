import { db, schema } from 'hub:db'
import { checkIncident } from '~~/server/utils/incident'

export default defineTask({
  meta: {
    name: 'incident',
    description: 'Handle incidents for all services',
  },
  async run() {
    // Fetch all services from database
    const services = await db.select().from(schema.services)

    for (const service of services) {
      await checkIncident(service)
    }

    return {
      result: 'success',
    }
  },
})
