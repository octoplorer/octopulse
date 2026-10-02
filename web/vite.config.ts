import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import VueRouter from 'vue-router/vite'
import UnoCSS from 'unocss/vite'
export default defineConfig({
  plugins: [UnoCSS(), VueRouter(), vue()],
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
