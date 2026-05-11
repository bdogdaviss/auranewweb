import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// The Go server in ../main.go embeds frontend/dist via go:embed and serves it at /, /products, /about.
// Dev server proxies API + auth routes to the running Go server so the React app talks to the real
// backend during development.
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    sourcemap: false,
  },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:3000',
      '/login': 'http://localhost:3000',
      '/register': 'http://localhost:3000',
      '/logout': 'http://localhost:3000',
      '/checkout': 'http://localhost:3000',
      '/download': 'http://localhost:3000',
      '/static': 'http://localhost:3000',
    },
  },
})
