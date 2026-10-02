import { readFile, writeFile } from 'node:fs/promises'
const document = JSON.parse(
  await readFile(new URL('../../api/openapi.json', import.meta.url), 'utf8'),
)
const sdkSource = await readFile(new URL('../src/client/sdk.gen.ts', import.meta.url), 'utf8')
const sdkNames = new Set([...sdkSource.matchAll(/export const (\w+)/g)].map((match) => match[1]))
const routes = []
for (const [path, item] of Object.entries(document.paths)) {
  for (const [method, operation] of Object.entries(item)) {
    if (!operation.operationId) continue
    if (!sdkNames.has(operation.operationId))
      throw new Error(`Missing HeyAPI SDK operation ${operation.operationId}`)
    const parameters = []
    const pattern = path
      .split('/')
      .map((segment) => {
        const match = /^\{(.+)\}$/.exec(segment)
        if (match) {
          parameters.push(match[1])
          return '([^/]+)'
        }
        return segment.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
      })
      .join('/')
    routes.push({
      method: method.toUpperCase(),
      pattern: `^${pattern}$`,
      parameters,
      operation: operation.operationId,
    })
  }
}
await writeFile(
  new URL('../src/client/routes.gen.ts', import.meta.url),
  `// Generated from Go OpenAPI by scripts/generate-routes.mjs. Do not edit.\nexport const routes = ${JSON.stringify(routes, null, 2)} as const\n`,
)
