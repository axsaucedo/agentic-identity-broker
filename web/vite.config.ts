import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// SPA fallback plugin: serve index.html for client-side routes
function spaSinglePageAppPlugin() {
  return {
    name: 'spa-fallback',
    configureServer(server) {
      return () => {
        server.middlewares.use((req, res, next) => {
          const url = req.url?.split('?')[0] || '';
          const hasFileExtension = /\.\w+$/.test(url);
          const isConsentPath = url.startsWith('/consent');

          // Serve index.html for consent routes without file extensions
          if (isConsentPath && !hasFileExtension && url !== '/consent/') {
            req.url = '/consent/index.html';
          }
          next();
        });
      };
    },
  };
}

export default defineConfig({
  plugins: [react(), spaSinglePageAppPlugin()],
  base: '/consent',
  build: {
    outDir: 'dist/consent',
    sourcemap: false,
    minify: 'terser',
  },
  server: {
    port: 3000,
    strictPort: false,
    open: false,
    proxy: {
      // Match all paths EXCEPT: node_modules, @vite, __vite, /consent (frontend files), and file extensions
      '^/(?!node_modules|@vite|__vite|consent).*': {
        target: process.env.VITE_API_URL || 'http://localhost:8000',
        changeOrigin: true,
        rewrite: (path) => path,
        configure: (proxy, _options) => {
          proxy.on('proxyReq', (proxyReq, req, res) => {
            // Add development principal for local testing
            // In production, authentication is handled by the browser/upstream proxy
            proxyReq.setHeader('X-Remote-User', 'dev@example.com');
          });
        },
      },
    },
  },
  resolve: {
    alias: {
      '@design-system': path.resolve(__dirname, './src/design-system'),
      '@components': path.resolve(__dirname, './src/components'),
      '@hooks': path.resolve(__dirname, './src/hooks'),
      '@services': path.resolve(__dirname, './src/services'),
      '@types': path.resolve(__dirname, './src/types'),
      '@utils': path.resolve(__dirname, './src/utils'),
      '@assets': path.resolve(__dirname, './src/assets'),
      '@styles': path.resolve(__dirname, './src/styles'),
    },
  },
})
