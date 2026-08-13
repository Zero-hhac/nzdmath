import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import path from 'path';
import {defineConfig, loadEnv} from 'vite';

export default defineConfig(({mode}) => {
  const env = loadEnv(mode, '.', '');
  const backendPort = env.VITE_BACKEND_PORT || '8080';
  const backendUrl = `http://localhost:${backendPort}`;
  return {
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, '.'),
      },
    },
    server: {
      port: 3000,
      host: '0.0.0.0',
      proxy: {
        '/api': {
          target: backendUrl,
          // 不能改 Origin/Host：后端 WS 握手做同源校验（Origin == Host），
          // changeOrigin:true 会把 Host 改成 8080，浏览器带 3000 的 Origin 时握手 403。
          changeOrigin: false,
          ws: true, // WebSocket（聊天室/通知推送通道）需要代理支持 Upgrade 握手
        },
        '/uploads': {
          target: backendUrl,
          changeOrigin: true,
        },
      },
      hmr: process.env.DISABLE_HMR !== 'true',
      watch: process.env.DISABLE_HMR === 'true' ? null : {},
    },
    build: {
      rollupOptions: {
        output: {
          manualChunks: {
            markdown: ['react-markdown', 'remark-gfm', 'rehype-raw', 'rehype-sanitize', 'katex', 'rehype-katex', 'remark-math'],
            charts: ['recharts'],
            docx: ['mammoth'],
          },
        },
      },
    },
  };
});