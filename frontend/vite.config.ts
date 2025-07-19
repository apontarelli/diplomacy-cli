import { defineConfig } from 'vite'
import tsconfigPaths from 'vite-tsconfig-paths'

export default defineConfig({
  plugins: [tsconfigPaths()],
  build: { 
    target: 'esnext',
    outDir: 'dist',
    assetsDir: 'assets'
  },
  server: { 
    port: 3000,
    proxy: { 
      '/api': 'http://localhost:8080',
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true
      }
    } 
  },
  resolve: {
    alias: {
      '@': '/src'
    }
  }
})