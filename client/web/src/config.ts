// Client configuration
// Reads from window.APP_CONFIG which is set by:
// - public/config.js for local development
// - docker-entrypoint.sh for production (generated at container startup)

// Type declaration for window.APP_CONFIG
declare global {
  interface Window {
    APP_CONFIG?: {
      API_BASE_URL?: string
    }
  }
}

// Get API base URL from runtime config, fallback to same-origin with port 8080
const apiBaseUrl = window.APP_CONFIG?.API_BASE_URL || `${window.location.protocol}//${window.location.hostname}:8080`

// Parse the URL to construct WebSocket URL
function buildConfig(baseUrl: string) {
  // Check if it's a relative path (e.g., "/api")
  if (baseUrl.startsWith('/')) {
    // Relative path - use current origin
    const httpBase = `${window.location.origin}${baseUrl}`
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsBase = `${wsProtocol}//${window.location.host}${baseUrl}/ws`
    return { httpBaseUrl: httpBase, wsUrl: wsBase }
  }

  // Absolute URL - parse it
  const url = new URL(baseUrl)
  const wsProtocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${wsProtocol}//${url.host}${url.pathname}/ws`.replace(/\/+ws$/, '/ws')

  return {
    httpBaseUrl: baseUrl.replace(/\/$/, ''), // Remove trailing slash
    wsUrl,
  }
}

export const config = buildConfig(apiBaseUrl)

// Log config on startup for debugging
console.log('BounceBot config:', config)
