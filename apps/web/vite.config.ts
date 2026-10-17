import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

// The browser talks to /api/* on the same origin as the page. Vite forwards
// those requests to the Go API, so there is no CORS and the session cookie
// just works. Set DEVFLOW_API to point at a different API address.
const apiTarget = process.env.DEVFLOW_API ?? 'http://localhost:8080'

const proxy = {
  '/api': {
    target: apiTarget,
    changeOrigin: true,
    rewrite: (path: string) => path.replace(/^\/api/, ''),
  },
}

export default defineConfig({
  plugins: [react()],
  server: { host: true, port: 5173, allowedHosts: true, proxy },
  preview: { host: true, port: 4173, allowedHosts: true, proxy },
  test: { environment: 'node' },
})
