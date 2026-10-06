import vue from '@vitejs/plugin-vue'
import UnoCSS from 'unocss/vite'
import { defineConfig } from 'vitest/config'
import VueRouter from 'vue-router/vite'

export default defineConfig({
  resolve: {
    tsconfigPaths: true,
  },
  plugins: [UnoCSS(), VueRouter(), vue()],
  test: {
    environment: 'node',
    maxWorkers: 4,
    include: ['src/**/*.{test,spec}.ts'],
  },
  server: {
    port: 5173,
    // The backend compares Origin with Host for write requests. Preserve the browser's authority.
    proxy: {
      '/api': { target: 'http://127.0.0.1:8080', changeOrigin: false },
      '/assets/uploads/': { target: 'http://127.0.0.1:8080', changeOrigin: false },
    },
  },
  build: { target: 'es2022', outDir: 'dist', emptyOutDir: true },
})
