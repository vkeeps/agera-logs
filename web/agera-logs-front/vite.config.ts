import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:9302', // 后端API服务地址
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''), // 把/api前缀去掉
        configure: (proxy) => {
          // 当代理错误时快速失败，避免长时间等待
          proxy.on('error', () => { /* 忽略错误日志 */ })
        },
        // 设置更短的超时时间，以便更快响应错误
        timeout: 10000,
      }
    }
  }
})
