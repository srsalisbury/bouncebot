// Client configuration
// Can be overridden via environment variables (prefixed with VITE_)

// Use current hostname so it works from other devices on the network
const serverHost = import.meta.env.VITE_SERVER_HOST || window.location.hostname || 'localhost'
const serverPort = import.meta.env.VITE_SERVER_PORT || '8080'

// Use same protocol as the page was loaded with (http/https)
const httpProtocol = window.location.protocol // 'http:' or 'https:'
const wsProtocol = httpProtocol === 'https:' ? 'wss:' : 'ws:'

// Optional base path for API calls (e.g., '/api' when behind a reverse proxy)
// Should NOT have a trailing slash
const basePath = (import.meta.env.VITE_API_BASE_PATH || '').replace(/\/$/, '')

// Construct base URLs
export const config = {
  // Base URL for HTTP/Connect-RPC calls (e.g., 'http://localhost:8080' or 'https://example.com/api')
  httpBaseUrl: `${httpProtocol}//${serverHost}:${serverPort}${basePath}`,

  // Base URL for WebSocket connections (e.g., 'ws://localhost:8080/ws' or 'wss://example.com/api/ws')
  wsUrl: `${wsProtocol}//${serverHost}:${serverPort}${basePath}/ws`,
}
