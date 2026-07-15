import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import VueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [VueDevTools(),vue()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080', // Your local Go backend
        changeOrigin: true,
        // This rewrite function does EXACTLY what your Nginx trailing slash does
        // It changes http://localhost:5173/api/users -> http://localhost:8080/users
        rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
})
