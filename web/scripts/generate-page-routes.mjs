import { fileURLToPath } from 'node:url'
import { createRoutesContext, resolveOptions } from 'vue-router/unplugin'

// Generate route types before vue-tsc, including on a clean checkout without a Vite server.
const context = createRoutesContext(
  resolveOptions({ root: fileURLToPath(new URL('../', import.meta.url)) }),
)
await context.scanPages(false)
