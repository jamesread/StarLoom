import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import basicSsl from '@vitejs/plugin-basic-ssl'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { createReadStream, existsSync } from 'node:fs'
import { stat } from 'node:fs/promises'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const backendPort = process.env.PORT || '8080'
const supplementalThemesDir = path.join(__dirname, 'node_modules/picocrank/supplemental-themes')

function contentTypeFor(filePath: string) {
  if (filePath.endsWith('.css')) return 'text/css; charset=utf-8'
  if (filePath.endsWith('.json')) return 'application/json; charset=utf-8'
  if (filePath.endsWith('.png')) return 'image/png'
  if (filePath.endsWith('.jpg') || filePath.endsWith('.jpeg')) return 'image/jpeg'
  if (filePath.endsWith('.webp')) return 'image/webp'
  if (filePath.endsWith('.svg')) return 'image/svg+xml'
  return 'application/octet-stream'
}

function copyDirectoryRecursive(sourceDir: string, targetDir: string) {
  fs.mkdirSync(targetDir, { recursive: true })

  for (const entry of fs.readdirSync(sourceDir, { withFileTypes: true })) {
    const sourcePath = path.join(sourceDir, entry.name)
    const targetPath = path.join(targetDir, entry.name)

    if (entry.isDirectory()) {
      copyDirectoryRecursive(sourcePath, targetPath)
      continue
    }

    fs.copyFileSync(sourcePath, targetPath)
  }
}

function supplementalThemesPlugin() {
  return {
    name: 'starapp-supplemental-themes',
    configureServer(server: { middlewares: { use: Function } }) {
      if (!fs.existsSync(supplementalThemesDir)) {
        return
      }

      server.middlewares.use('/supplemental-themes', (req, res, next) => {
        const requestPath = decodeURIComponent((req.url || '/').split('?')[0])
        const relativePath = requestPath.replace(/^\/+/, '')
        const filePath = path.resolve(supplementalThemesDir, relativePath)

        if (!filePath.startsWith(path.resolve(supplementalThemesDir))) {
          res.statusCode = 403
          res.end('Forbidden')
          return
        }

        if (!existsSync(filePath)) {
          next()
          return
        }

        stat(filePath)
          .then((fileStat) => {
            if (!fileStat.isFile()) {
              next()
              return
            }

            res.setHeader('Content-Type', contentTypeFor(filePath))
            createReadStream(filePath).pipe(res)
          })
          .catch(() => {
            next()
          })
      })
    },
    closeBundle() {
      if (!fs.existsSync(supplementalThemesDir)) {
        return
      }

      const outDir = path.resolve(__dirname, 'dist', 'supplemental-themes')
      copyDirectoryRecursive(supplementalThemesDir, outDir)
    },
  }
}

export default defineConfig({
  plugins: [vue(), basicSsl(), supplementalThemesPlugin()],
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
