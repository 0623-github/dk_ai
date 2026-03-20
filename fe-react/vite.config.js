import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:9090',
        changeOrigin: true,
      },
      '/ping': {
        target: 'http://localhost:9090',
        changeOrigin: true,
      },
      '/chat': {
        target: 'http://localhost:9090',
        changeOrigin: true,
      }
    }
  }
})
