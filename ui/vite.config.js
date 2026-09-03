import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'

export default defineConfig({
  plugins: [vue(), vueJsx()],
  // The build output is embedded into the Go binary (src/internal/web).
  build: {
    outDir: '../src/internal/web/dist',
    // Keep .gitkeep in place: go:embed needs dist/ to hold at least one file
    // even before the UI has ever been built. `make ui` clears stale assets.
    emptyOutDir: false,
  },
  server: {
    port: 4000,
    host: true,
    open: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      '@': '/src'
    }
  }
})
