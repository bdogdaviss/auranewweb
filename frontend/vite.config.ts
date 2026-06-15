import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// The Go server in ../main.go embeds frontend/dist via go:embed and serves it at /, /products, /about.
// Dev server proxies API + auth routes (and legacy template pages) to the running Go server.
// /account is handled client-side by React (no proxy) so dev assets work correctly.

export default defineConfig({
  plugins: [react(), tailwindcss()],
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
      // Note: /account is now a client-side React route (AccountPage). Do not proxy it,
      // or the dev server will get the production-built index.html from Go while
      // expecting dev assets (causing 404s on JS/CSS). The /api proxy handles data calls.
      '/recover': 'http://localhost:3000',
      '/success': 'http://localhost:3000',
      '/auth': 'http://localhost:3000',
      '/static': 'http://localhost:3000',
    },
  },
})
