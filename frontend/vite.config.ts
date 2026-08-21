import { fileURLToPath, URL } from 'node:url'
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
// 后端地址等环境相关配置放在 .env*（见 .env.example），不再硬编码在这里。
export default defineConfig(({ command, mode }) => {
  // 第三个参数 ''：加载全部变量（默认只暴露 VITE_ 前缀）
  const env = loadEnv(mode, process.cwd(), '')

  return {
    // build 时资源 base 用 /resources/（与后端 embed 的 /resources/* 静态服务对齐，
    // 见 mars/frontend/embed.go）；开发模式保持 / 由 vite dev server 直接服务。
    base: command === 'build' ? '/resources/' : '/',
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    build: {
      // 产物目录用 build/（默认 dist/）：并入 mars 后由 frontend/embed.go 的
      // //go:embed build/* 内嵌到 /resources/，与旧版 CRA 产物目录契约一致，
      // Makefile build_web / Dockerfile / CI 无需改动。
      outDir: 'build',
      // CodeMirror 懒 chunk（~600KB）是刻意按需加载的编辑器依赖，非误打包，
      // 阈值提到 700KB 让这类合理大 chunk 不再触发 >500kB 警告。
      chunkSizeWarningLimit: 700,
      rollupOptions: {
        output: {
          // 第三方框架拆成稳定 vendor chunk：react 版本不动时 hash 不变，
          // 浏览器跨部署复用缓存，避免每次改代码都让 react 全家桶跟着重新下载。
          manualChunks: {
            'vendor-react': ['react', 'react-dom', 'react-router-dom'],
          },
        },
      },
    },
    server: {
      port: 5173,
      host: true,
      proxy: {
        // 开发期把 /api 代理到 VITE_API_TARGET，免 CORS
        '/api': {
          target: env.VITE_API_TARGET,
          changeOrigin: true,
          secure: true,
        },
        // WebSocket 实时通道：/ws 代理到 VITE_WS_TARGET，用于终端/部署/集群信息推送。
        // changeOrigin 必须配：改写 Host 为后端域名，否则握手转发失败（Empty reply）。
        '/ws': {
          target: env.VITE_WS_TARGET,
          changeOrigin: true,
          ws: true,
          secure: true,
        },
      },
    },
  }
})
