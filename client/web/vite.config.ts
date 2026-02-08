import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  base: process.env.VITE_BASE_PATH || '/',
  plugins: [vue()],
  define: {
    __APP_VERSION__: JSON.stringify(process.env.VITE_APP_VERSION || 'dev'),
  },
  server: {
    host: true, // Listen on all network interfaces
    allowedHosts: true, // Allow any hostname (for local network access)
  },
  test: {
    environment: 'jsdom',
  },
})
