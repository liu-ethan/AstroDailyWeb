import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    host: '0.0.0.0',
    port: 9091, // 正式地址
    // port: 19091,  // 测试地址
    proxy: {
      '/api': {
        target: 'http://localhost:9090', // 正式地址
        // target: 'http://127.0.0.1:19090', // 测试地址
        changeOrigin: true,
      },
    },
  },
})
