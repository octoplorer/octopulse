import assert from 'node:assert/strict'
import { createServer as createHTTPServer, request } from 'node:http'
import { once } from 'node:events'
import { fileURLToPath } from 'node:url'
import test from 'node:test'
import { createServer, loadConfigFromFile } from 'vite'

function proxyRequest(port, path, headers) {
  return new Promise((resolve, reject) => {
    const outgoing = request(
      {
        hostname: '127.0.0.1',
        port,
        path,
        method: path.startsWith('/api/') ? 'POST' : 'GET',
        headers,
      },
      async (response) => {
        try {
          const chunks = []
          for await (const chunk of response) chunks.push(chunk)
          resolve({
            status: response.statusCode,
            headers: response.headers,
            body: JSON.parse(Buffer.concat(chunks).toString()),
          })
        } catch (error) {
          reject(error)
        }
      },
    )
    outgoing.on('error', reject)
    outgoing.end()
  })
}

test('development proxy preserves the browser authority and security headers', async (t) => {
  const backend = createHTTPServer((request, response) => {
    response.setHeader('Content-Type', 'application/json')
    response.setHeader(
      'Set-Cookie',
      'octopulse_session=test-session; Path=/api/v1; HttpOnly; SameSite=Strict',
    )
    response.end(
      JSON.stringify({
        host: request.headers.host,
        origin: request.headers.origin,
        cookie: request.headers.cookie,
        csrf: request.headers['x-csrf-token'],
      }),
    )
  })
  backend.listen(0, '127.0.0.1')
  await once(backend, 'listening')
  t.after(() => new Promise((resolve) => backend.close(resolve)))
  const target = `http://127.0.0.1:${backend.address().port}`
  const root = fileURLToPath(new URL('../', import.meta.url))
  const loaded = await loadConfigFromFile(
    { command: 'serve', mode: 'development' },
    fileURLToPath(new URL('../vite.config.ts', import.meta.url)),
  )
  assert.ok(loaded)
  // Keep the real proxy options; only substitute an isolated backend port.
  const proxy = Object.fromEntries(
    Object.entries(loaded.config.server.proxy).map(([path, options]) => [
      path,
      typeof options === 'string' ? target : { ...options, target },
    ]),
  )
  const vite = await createServer({
    configFile: false,
    root,
    plugins: [],
    optimizeDeps: { noDiscovery: true, include: [] },
    server: { host: '127.0.0.1', port: 0, proxy },
  })
  t.after(() => vite.close())
  await vite.listen()
  const port = vite.httpServer.address().port

  for (const hostname of ['127.0.0.1', 'localhost']) {
    for (const path of ['/api/v1/session', '/api/v1/setup', '/assets/uploads/logo.png']) {
      await t.test(`${hostname}${path}`, async () => {
        const host = `${hostname}:${port}`
        const origin = `http://${host}`
        const response = await proxyRequest(port, path, {
          Host: host,
          Origin: origin,
          Cookie: 'octopulse_session=test-session',
          'X-CSRF-Token': 'test-csrf',
        })
        assert.equal(response.status, 200)
        assert.deepEqual(response.body, {
          host,
          origin,
          cookie: 'octopulse_session=test-session',
          csrf: 'test-csrf',
        })
        assert.match(response.headers['set-cookie'][0], /SameSite=Strict/)
      })
    }
  }
  await t.test('foreign origins remain visible to the backend', async () => {
    const response = await proxyRequest(port, '/api/v1/session', {
      Origin: 'https://untrusted.example',
    })
    assert.equal(response.body.origin, 'https://untrusted.example')
  })
})
