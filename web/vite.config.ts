import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  // history 路由需要绝对根路径，相对 base 会让 /notes/xxx 下的静态资源 404。
  base: '/',
  plugins: [vue()],
  build: {
    // 产物直接进入 Go embed 目录，构建脚本随后用 -tags embedui 编译。
    outDir: '../app/internal/webui/dist',
    emptyOutDir: true,
    chunkSizeWarningLimit: 800
  },
  server: {
    // 源码位于 /mnt/e（Windows drvfs 挂载），inotify 事件不会传递到 WSL，
    // 不开启轮询会导致改完代码 Vite 仍返回旧编译结果。轮询间隔取 300ms 折中 CPU 占用。
    watch: { usePolling: true, interval: 300 },
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true
      }
    }
  }
})
