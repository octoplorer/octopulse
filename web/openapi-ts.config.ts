import { defineConfig } from '@hey-api/openapi-ts'
export default defineConfig({
  input: '../api/openapi.json',
  output: 'src/client',
  plugins: [
    '@hey-api/typescript',
    '@hey-api/client-fetch',
    '@hey-api/sdk',
    { name: '@pinia/colada', queryOptions: true },
  ],
})
