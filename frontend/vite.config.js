import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      '^/(account|video|feed|like|comment|follow|message)(/|$)': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/static/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
});