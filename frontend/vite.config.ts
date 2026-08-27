import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import basicSsl from '@vitejs/plugin-basic-ssl'

const backendPort = process.env.PORT || '8080'

export default defineConfig({
  plugins: [vue(), basicSsl()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: `http://localhost:${backendPort}`,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '') || '/',
      },
      '/avatars': {
        target: `http://localhost:${backendPort}`,
        changeOrigin: true,
      },
    },
  },
})
