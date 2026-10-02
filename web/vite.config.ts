import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import UnoCSS from 'unocss/vite'
export default defineConfig({
  plugins: [UnoCSS(), vue()],
  server: {
    port: 5173,
    proxy: { '/api': 'http://127.0.0.1:8080', '/assets/uploads/': 'http://127.0.0.1:8080' },
  },
  build: { target: 'es2022', outDir: 'dist', emptyOutDir: true },
})
