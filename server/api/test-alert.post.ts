/**
 * POST /test-alert
 * Test alert notification (for debugging)
 */
export default defineEventHandler(async (event) => {
  // Verify authentication
  verifyAuth(event);

  const body = await readBody(event);
  
  // This is a placeholder for testing
  console.log('Test alert triggered:', {
    message: body.message || 'Test alert',
    timestamp: new Date().toISOString(),
  });

  return {
    success: true,
    message: 'Test alert sent (placeholder)',
    timestamp: new Date().toISOString(),
  };
});
