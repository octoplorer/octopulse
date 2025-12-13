export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()

  // Vercel CRON_SECRET security check
  if (getHeader(event, 'authorization') !== `Bearer ${config.cronSecret}`) {
    throw createError({ statusCode: 401 })
  }

  // Call Nitro task
  const { result } = await runTask('check')
  return { success: true, result }
})
