import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'http://ttt_backend:8080', // <<< ВАЖНО! НЕ localhost
        changeOrigin: true,
        secure: false,
      },
      '/ws': {
        target: 'ws://ttt_backend:8080',
        ws: true,
        changeOrigin: true,
      },
    }
  }
})
