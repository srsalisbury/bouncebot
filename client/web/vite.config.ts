import { defineConfig, type Plugin } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import { writeFileSync } from 'fs'
import { resolve } from 'path'

function versionFilePlugin(): Plugin {
  const version = process.env.VITE_APP_VERSION || 'dev'
  return {
    name: 'version-file',
    writeBundle(options) {
      const dir = options.dir || resolve(__dirname, 'dist')
      writeFileSync(resolve(dir, 'version.json'), JSON.stringify({ version }))
    },
  }
}

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), versionFilePlugin()],
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
