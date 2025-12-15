import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    host: true, // Listen on all network interfaces
    allowedHosts: true, // Allow any hostname (for local network access)
  },
})
