import { db, schema } from 'hub:db'; 

const monitors = schema.monitors;

/**
 * GET /monitors/public
 * Get public monitoring data (no authentication required)
 */
export default defineEventHandler(async (event) => {
  // Get all monitors with public information only
  const allMonitors = await db.select({
    id: monitors.id,
    name: monitors.name,
    url: monitors.url,
    status: monitors.status,
    lastCheck: monitors.lastCheck,
    certExpiry: monitors.certExpiry,
    domainExpiry: monitors.domainExpiry,
  }).from(monitors);

  // Calculate overall system status
  const hasDown = allMonitors.some(m => m.status === 'DOWN');
  const systemStatus = hasDown ? 'Disruption' : 'Operational';

  return {
    status: systemStatus,
    monitors: allMonitors,
  };
});
